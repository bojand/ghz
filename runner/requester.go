package runner

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/bojand/ghz/load"
	"github.com/bojand/ghz/protodesc"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"

	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	// To register the xds resolvers and balancers.
	_ "google.golang.org/grpc/xds"
)

// Max size of the buffer of result channel.
const maxResult = 1000000

// result of a call
type callResult struct {
	err       error
	status    string
	duration  time.Duration
	timestamp time.Time
}

// Requester is used for doing the requests
type Requester struct {
	conns    []*grpc.ClientConn
	stubs    []grpcdynamic.Stub
	handlers []*statsHandler

	mtd      *desc.MethodDescriptor
	reporter *Reporter

	config *RunConfig

	results chan *callResult
	stopCh  chan bool
	start   time.Time

	dataProvider     DataProviderFunc
	metadataProvider MetadataProviderFunc

	lock       sync.Mutex
	stopReason StopReason
	workers    []*Worker
}

// NewRequester creates a new requestor from the passed RunConfig
func NewRequester(c *RunConfig) (*Requester, error) {

	var err error
	var mtd *desc.MethodDescriptor

	reqr := &Requester{
		config:     c,
		stopReason: ReasonNormalEnd,
		results:    make(chan *callResult, min(c.c*1000, maxResult)),
		stopCh:     make(chan bool, 1),
		workers:    make([]*Worker, 0, c.c),
		conns:      make([]*grpc.ClientConn, 0, c.nConns),
		stubs:      make([]grpcdynamic.Stub, 0, c.nConns),
	}

	if c.proto != "" {
		mtd, err = protodesc.GetMethodDescFromProto(c.call, c.proto, c.importPaths)
	} else if c.protoset != "" {
		mtd, err = protodesc.GetMethodDescFromProtoSet(c.call, c.protoset)
	} else if c.protosetBinary != nil {
		mtd, err = protodesc.GetMethodDescFromProtoSetBinary(c.call, c.protosetBinary)
	} else {
		// use reflection to get method descriptor
		var cc *grpc.ClientConn
		// temporary connection for reflection, do not store as requester connections
		cc, err = reqr.newClientConn(false)
		if err != nil {
			return nil, err
		}

		defer func() {
			// purposefully ignoring error as we do not care if there
			// is an error on close
			_ = cc.Close()
		}()

		// cancel is ignored here as connection.Close() is used.
		// See https://godoc.org/google.golang.org/grpc#DialContext
		ctx, _ := context.WithTimeout(context.Background(), c.dialTimeout)

		md := make(metadata.MD)
		if c.rmd != nil && len(c.rmd) > 0 {
			md = metadata.New(c.rmd)
		}

		refCtx := metadata.NewOutgoingContext(ctx, md)

		refClient := grpcreflect.NewClientAuto(refCtx, cc)

		mtd, err = protodesc.GetMethodDescFromReflect(c.call, refClient)
	}

	if err != nil {
		return nil, err
	}

	md := mtd.GetInputType()
	payloadMessage := dynamic.NewMessage(md)
	if payloadMessage == nil {
		return nil, fmt.Errorf("no input type of method: %s", mtd.GetName())
	}

	// fill in the rest
	reqr.mtd = mtd

	if c.dataProviderFunc != nil {
		reqr.dataProvider = c.dataProviderFunc
	} else {
		defaultDataProvider, err := newDataProvider(reqr.mtd, c.binary, c.dataFunc, c.data, !c.disableTemplateFuncs, !c.disableTemplateData, c.funcs)
		if err != nil {
			return nil, err
		}
		reqr.dataProvider = defaultDataProvider.getDataForCall
	}

	if c.mdProviderFunc != nil {
		reqr.metadataProvider = c.mdProviderFunc
	} else {
		defaultMDProvider, err := newMetadataProvider(reqr.mtd, c.metadata, !c.disableTemplateFuncs, !c.disableTemplateData, c.funcs)
		if err != nil {
			return nil, err
		}
		reqr.metadataProvider = defaultMDProvider.getMetadataForCall
	}

	return reqr, nil
}

// Run makes all the requests and returns a report of results
// It blocks until all work is done.
func (b *Requester) Run() (*Report, error) {

	defer close(b.stopCh)

	cc, err := b.openClientConns()
	if err != nil {
		return nil, err
	}

	start := time.Now()

	b.lock.Lock()
	b.start = start

	// create a client stub for each connection
	for n := 0; n < b.config.nConns; n++ {
		stub := grpcdynamic.NewStub(cc[n])
		b.stubs = append(b.stubs, stub)
	}

	b.reporter = newReporter(b.results, b.config)
	b.lock.Unlock()

	go func() {
		b.reporter.Run()
	}()

	wt := createWorkerTicker(b.config)

	p := createPacer(b.config)

	err = b.runWorkers(wt, p)

	report := b.Finish()

	b.closeClientConns()

	return report, err
}

// Stop stops the test
func (b *Requester) Stop(reason StopReason) {

	b.stopCh <- true

	b.lock.Lock()
	b.stopReason = reason

	if b.config.hasLog {
		b.config.log.Debugf("Stopping with reason: %+v", reason)
	}
	b.lock.Unlock()

	if b.config.zstop == "close" {
		b.closeClientConns()
	} else if b.config.zstop == "ignore" {
		for _, h := range b.handlers {
			h.Ignore(true)
		}
		b.closeClientConns()
	}
}

// Finish finishes the test run
func (b *Requester) Finish() *Report {
	close(b.results)
	total := time.Since(b.start)

	if b.config.hasLog {
		b.config.log.Debug("Waiting for report")
	}

	// Wait until the reporter is done.
	<-b.reporter.done

	if b.config.hasLog {
		b.config.log.Debug("Finalizing report")
	}

	var r StopReason
	b.lock.Lock()
	r = b.stopReason
	b.lock.Unlock()

	return b.reporter.Finalize(r, total)
}

func (b *Requester) openClientConns() ([]*grpc.ClientConn, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.conns) == b.config.nConns {
		return b.conns, nil
	}

	for n := 0; n < b.config.nConns; n++ {
		c, err := b.newClientConn(true)
		if err != nil {
			if b.config.hasLog {
				b.config.log.Errorf("Error creating client connection: %+v", err.Error())
			}

			return nil, err
		}

		b.conns = append(b.conns, c)
	}

	return b.conns, nil
}

func (b *Requester) closeClientConns() {
	if b.config.hasLog {
		b.config.log.Debug("Closing client connections")
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	if b.conns == nil {
		return
	}

	for _, cc := range b.conns {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
		defer cancel()

		shutdownCh := connectionOnState(ctx, cc, connectivity.Shutdown)

		_ = cc.Close()

		<-shutdownCh
	}

	b.conns = nil
}

func (b *Requester) newClientConn(withStatsHandler bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if b.config.insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(b.config.creds))
	}

	if b.config.authority != "" {
		opts = append(opts, grpc.WithAuthority(b.config.authority))
	}

	if len(b.config.defaultCallOptions) > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(b.config.defaultCallOptions...))
	} else {
		// increase max receive and send message sizes
		opts = append(opts,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(math.MaxInt32),
				grpc.MaxCallSendMsgSize(math.MaxInt32),
			))

	}

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, b.config.dialTimeout)
	// cancel is ignored here as connection.Close() is used.
	// See https://godoc.org/google.golang.org/grpc#DialContext

	if b.config.keepaliveTime > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    b.config.keepaliveTime,
			Timeout: b.config.keepaliveTime,
		}))
	}

	if withStatsHandler {
		sh := &statsHandler{
			id:      len(b.handlers),
			results: b.results,
			hasLog:  b.config.hasLog,
			log:     b.config.log,
		}

		b.handlers = append(b.handlers, sh)

		opts = append(opts, grpc.WithStatsHandler(sh))
	}

	if b.config.hasLog {
		b.config.log.Debugw("Creating client connection", "options", opts)
	}

	if b.config.lbStrategy != "" {
		grpcServiceConfig := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, b.config.lbStrategy)
		opts = append(opts, grpc.WithDefaultServiceConfig(grpcServiceConfig))
	}

	// create client connection
	return grpc.DialContext(ctx, b.config.host, opts...)
}

func (b *Requester) runWorkers(wt load.WorkerTicker, p load.Pacer) error {

	wct := wt.Ticker()

	var wm sync.Mutex

	// worker control ticker goroutine
	go func() {
		wt.Run()
	}()

	errC := make(chan error, b.config.c)
	done := make(chan struct{})
	workerTickerDone := make(chan struct{})
	ticks := make(chan TickValue)
	counter := Counter{}

	go func() {
		defer close(workerTickerDone)
		n := 0
		wc := 0
		for tv := range wct {
			if b.config.hasLog {
				b.config.log.Debugw("Worker ticker.", "delta", tv.Delta)
			}

			if tv.Delta > 0 {
				for i := 0; i < tv.Delta; i++ {
					wID := "g" + strconv.Itoa(wc) + "c" + strconv.Itoa(n)

					if len(b.config.name) > 0 {
						wID = b.config.name + ":" + wID
					}

					if b.config.hasLog {
						b.config.log.Debugw("Creating worker with ID: "+wID, "workerID", wID)
					}

					w := Worker{
						ticks:            ticks,
						active:           true,
						stub:             b.stubs[n],
						mtd:              b.mtd,
						config:           b.config,
						stopCh:           make(chan bool),
						workerID:         wID,
						dataProvider:     b.dataProvider,
						metadataProvider: b.metadataProvider,
						streamRecv:       b.config.recvMsgFunc,
						msgProvider:      b.config.dataStreamFunc,
					}

					wc++ // increment worker id

					n++ // increment connection counter

					// wrap around connections if needed
					if n == b.config.nConns {
						n = 0
					}

					wm.Lock()
					b.workers = append(b.workers, &w)
					wm.Unlock()

					go func() {
						errC <- w.runWorker()
					}()
				}
			} else if tv.Delta < 0 {
				nd := -1 * tv.Delta
				wm.Lock()
				wdc := 0
				for _, wrk := range b.workers {
					if wdc == nd {
						break
					}

					wrk := wrk
					if wrk.active {
						wrk.Stop()
						wdc++
					}
				}
				wm.Unlock()
			}
			if tv.Done {
				return
			}
		}
	}()

	go func() {
		defer close(ticks)
		defer func() {
			wm.Lock()
			nw := len(b.workers)
			for i := 0; i < nw; i++ {
				b.workers[i].Stop()
			}
			wm.Unlock()
		}()

		defer func() {
			wt.Finish()
			<-workerTickerDone
		}()

		began := time.Now()

		for {
			wait, stop := p.Pace(time.Since(began), counter.Get())

			if stop {
				if b.config.hasLog {
					b.config.log.Debugw("Received stop from pacer.")
				}
				done <- struct{}{}
				return
			}

			if wait > 0 {
				time.Sleep(wait)
			}

			select {
			case ticks <- TickValue{instant: time.Now(), reqNumber: counter.Inc() - 1}:
				continue
			case <-b.stopCh:
				if b.config.hasLog {
					b.config.log.Debugw("Signal received from stop channel.", "count", counter.Get())
				}
				done <- struct{}{}
				return
			}
		}
	}()

	<-done

	var err error
	wm.Lock()
	nw := len(b.workers)
	wm.Unlock()
	for i := 0; i < nw; i++ {
		err = multierr.Append(err, <-errC)
	}

	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func createWorkerTicker(config *RunConfig) load.WorkerTicker {
	if config.workerTicker != nil {
		return config.workerTicker
	}

	var wt load.WorkerTicker
	switch config.cSchedule {
	case ScheduleLine:
		wt = &load.LineWorkerTicker{
			C:           make(chan load.TickValue),
			Start:       config.cStart,
			Slope:       config.cStep,
			Stop:        config.cEnd,
			MaxDuration: config.cMaxDuration,
		}
	case ScheduleStep:
		wt = &load.StepWorkerTicker{
			C:            make(chan load.TickValue),
			Start:        config.cStart,
			Step:         config.cStep,
			Stop:         config.cEnd,
			StepDuration: config.cStepDuration,
			MaxDuration:  config.cMaxDuration,
		}
	default:
		wt = &load.ConstWorkerTicker{N: uint(config.c), C: make(chan load.TickValue)}
	}

	return wt
}

func createPacer(config *RunConfig) load.Pacer {
	if config.pacer != nil {
		return config.pacer
	}

	var p load.Pacer
	switch config.loadSchedule {
	case ScheduleLine:
		p = &load.LinearPacer{
			Start:        load.ConstantPacer{Freq: uint64(config.loadStart), Max: uint64(config.n)},
			Slope:        int64(config.loadStep),
			Stop:         load.ConstantPacer{Freq: uint64(config.loadEnd), Max: uint64(config.n)},
			LoadDuration: config.loadDuration,
			Max:          uint64(config.n),
		}
	case ScheduleStep:
		p = &load.StepPacer{
			Start:        load.ConstantPacer{Freq: uint64(config.loadStart), Max: uint64(config.n)},
			Step:         int64(config.loadStep),
			Stop:         load.ConstantPacer{Freq: uint64(config.loadEnd), Max: uint64(config.n)},
			LoadDuration: config.loadDuration,
			StepDuration: config.loadStepDuration,
			Max:          uint64(config.n),
		}
	default:
		p = &load.ConstantPacer{Freq: uint64(config.rps), Max: uint64(config.n)}
	}

	return p
}

func checkState(conn *grpc.ClientConn, states ...connectivity.State) bool {
	currentState := conn.GetState()
	for _, s := range states {
		if currentState == s {
			return true
		}
	}

	return false
}

func connectionOnState(ctx context.Context, conn *grpc.ClientConn, states ...connectivity.State) <-chan bool {

	stateCh := make(chan bool)

	go func() {
		defer close(stateCh)
		if checkState(conn, states...) {
			stateCh <- true
			return
		}

		for {
			change := conn.WaitForStateChange(ctx, conn.GetState())
			if !change {
				stateCh <- checkState(conn, states...)
				return
			}

			if checkState(conn, states...) {
				stateCh <- true
				return
			}
		}
	}()

	return stateCh
}
