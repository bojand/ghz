package runner

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bojand/ghz/protodesc"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"

	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// Max size of the buffer of result channel.
const maxResult = 1000000

// result of a call
type callResult struct {
	err      error
	status   string
	duration time.Duration
}

// Requester is used for doing the requests
type Requester struct {
	cc    []*grpc.ClientConn
	stubs []grpcdynamic.Stub

	mtd      *desc.MethodDescriptor
	reporter *Reporter

	config  *RunConfig
	results chan *callResult
	stopCh  chan bool
	start   time.Time

	qpsTick time.Duration

	reqCounter int64

	stopReason StopReason
	lock       sync.Mutex
}

func newRequester(c *RunConfig) (*Requester, error) {
	var err error
	var mtd *desc.MethodDescriptor

	var qpsTick time.Duration
	if c.qps > 0 {
		qpsTick = time.Duration(1e6/(c.qps)) * time.Microsecond
	}

	reqr := &Requester{
		config:     c,
		qpsTick:    qpsTick,
		stopReason: ReasonNormalEnd,
		results:    make(chan *callResult, min(c.c*1000, maxResult)),
		stopCh:     make(chan bool, c.c),
		cc:         make([]*grpc.ClientConn, 0, c.nConns),
		stubs:      make([]grpcdynamic.Stub, 0, c.nConns),
	}

	// TODO REMOVE
	if reqr.config.nConns <= 0 {
		reqr.config.nConns = 5
	}

	if c.proto != "" {
		mtd, err = protodesc.GetMethodDescFromProto(c.call, c.proto, c.importPaths)
	} else if c.protoset != "" {
		mtd, err = protodesc.GetMethodDescFromProtoSet(c.call, c.protoset)
	} else {
		// use reflection to get method decriptor
		var cc *grpc.ClientConn
		// temporary connection for reflection, do not store as requester.cc
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
		if c.rmd != nil && len(*c.rmd) > 0 {
			md = metadata.New(*c.rmd)
		}

		refCtx := metadata.NewOutgoingContext(ctx, md)

		refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))

		mtd, err = protodesc.GetMethodDescFromReflect(c.call, refClient)
	}

	if err != nil {
		return nil, err
	}

	md := mtd.GetInputType()
	payloadMessage := dynamic.NewMessage(md)
	if payloadMessage == nil {
		return nil, fmt.Errorf("No input type of method: %s", mtd.GetName())
	}

	// fill in the rest
	reqr.mtd = mtd

	return reqr, nil
}

// Run makes all the requests and returns a report of results
// It blocks until all work is done.
func (b *Requester) Run() (*Report, error) {
	start := time.Now()

	cc, err := b.openClientConns()
	if err != nil {
		return nil, err
	}

	b.lock.Lock()
	b.start = start

	for n := 0; n < b.config.nConns; n++ {
		stub := grpcdynamic.NewStub(cc[n])
		b.stubs = append(b.stubs, stub)
	}

	b.reporter = newReporter(b.results, b.config)
	b.lock.Unlock()

	go func() {
		b.reporter.Run()
	}()

	err = b.runWorkers()

	report := b.Finish()
	b.closeClientConns()

	return report, err
}

// Stop stops the test
func (b *Requester) Stop(reason StopReason) {
	// Send stop signal so that workers can stop gracefully.
	for i := 0; i < b.config.c; i++ {
		b.stopCh <- true
	}

	b.lock.Lock()
	b.stopReason = reason
	b.lock.Unlock()

	b.closeClientConns()
}

// Finish finishes the test run
func (b *Requester) Finish() *Report {
	close(b.results)
	total := time.Now().Sub(b.start)

	// Wait until the reporter is done.
	<-b.reporter.done

	return b.reporter.Finalize(b.stopReason, total)
}

func (b *Requester) openClientConns() ([]*grpc.ClientConn, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.cc) == b.config.nConns {
		return b.cc, nil
	}

	for n := 0; n < b.config.nConns; n++ {
		c, err := b.newClientConn(true)
		if err != nil {
			return nil, err
		}

		b.cc = append(b.cc, c)
	}

	return b.cc, nil
}

func (b *Requester) closeClientConns() {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.cc == nil {
		return
	}

	for _, cc := range b.cc {
		_ = cc.Close()
	}

	b.cc = nil
}

func (b *Requester) newClientConn(withStatsHandler bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if b.config.insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(b.config.creds))
	}

	if b.config.authority != "" {
		opts = append(opts, grpc.WithAuthority(b.config.authority))
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
		opts = append(opts, grpc.WithStatsHandler(&statsHandler{b.results}))
	}

	// create client connection
	return grpc.DialContext(ctx, b.config.host, opts...)
}

func (b *Requester) runWorkers() error {
	nReqPerWorker := b.config.n / b.config.c

	if b.config.c == 0 {
		return nil
	}

	errC := make(chan error, b.config.c)

	n := 0 // connection counter
	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.config.c; i++ {

		wID := "g" + strconv.Itoa(i) + "c" + strconv.Itoa(n)
		w := Worker{
			stub:       b.stubs[n],
			mtd:        b.mtd,
			config:     b.config,
			stopCh:     b.stopCh,
			qpsTick:    b.qpsTick,
			reqCounter: &b.reqCounter,
			nReq:       nReqPerWorker,
			workerID:   wID,
		}

		n++ // increment connection counter

		// wrap around if needed
		if n == b.config.nConns {
			n = 0
		}

		go func() {
			errC <- w.runWorker()
		}()
	}

	var err error
	for i := 0; i < b.config.c; i++ {
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
