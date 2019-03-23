package runner

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
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
	cc       *grpc.ClientConn
	stub     grpcdynamic.Stub
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

	cc, err := b.openClientConn()
	if err != nil {
		return nil, err
	}

	b.lock.Lock()
	b.start = start
	b.stub = grpcdynamic.NewStub(cc)
	b.reporter = newReporter(b.results, b.config)
	b.lock.Unlock()

	go func() {
		b.reporter.Run()
	}()

	err = b.runWorkers()

	report := b.Finish()
	b.closeClientConn()

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

	b.closeClientConn()
}

// Finish finishes the test run
func (b *Requester) Finish() *Report {
	close(b.results)
	total := time.Now().Sub(b.start)

	// Wait until the reporter is done.
	<-b.reporter.done

	return b.reporter.Finalize(b.stopReason, total)
}

func (b *Requester) openClientConn() (*grpc.ClientConn, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.cc != nil {
		return b.cc, nil
	}
	cc, err := b.newClientConn(true)
	if err != nil {
		return nil, err
	}
	b.cc = cc
	return b.cc, nil
}

func (b *Requester) closeClientConn() {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.cc == nil {
		return
	}
	_ = b.cc.Close()
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
	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.config.c; i++ {
		go func() {
			errC <- b.runWorker(nReqPerWorker)
		}()
	}

	var err error
	for i := 0; i < b.config.c; i++ {
		err = multierr.Append(err, <-errC)
	}
	return err
}

func (b *Requester) runWorker(n int) error {
	var throttle <-chan time.Time
	if b.config.qps > 0 {
		throttle = time.Tick(b.qpsTick)
	}

	var err error
	for i := 0; i < n; i++ {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-b.stopCh:
			return nil
		default:
			if b.config.qps > 0 {
				<-throttle
			}
			err = multierr.Append(err, b.makeRequest())
		}
	}
	return err
}

func (b *Requester) makeRequest() error {

	reqNum := atomic.AddInt64(&b.reqCounter, 1)

	ctd := newCallTemplateData(b.mtd, reqNum)

	var input *dynamic.Message
	var streamInput *[]*dynamic.Message

	if !b.config.binary {
		data, err := ctd.executeData(string(b.config.data))
		if err != nil {
			return err
		}
		input, streamInput, err = createPayloads(string(data), b.mtd)
		if err != nil {
			return err
		}
	} else {
		var err error
		input, streamInput, err = createPayloadsFromBin(b.config.data, b.mtd)
		if err != nil {
			return err
		}
	}

	mdMap, err := ctd.executeMetadata(string(b.config.metadata))
	if err != nil {
		return err
	}

	var reqMD *metadata.MD
	if mdMap != nil && len(*mdMap) > 0 {
		md := metadata.New(*mdMap)
		reqMD = &md
	}

	ctx := context.Background()
	var cancel context.CancelFunc

	if b.config.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, b.config.timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// include the metadata
	if reqMD != nil {
		ctx = metadata.NewOutgoingContext(ctx, *reqMD)
	}

	// RPC errors are handled via stats handler

	if b.mtd.IsClientStreaming() && b.mtd.IsServerStreaming() {
		_ = b.makeBidiRequest(&ctx, streamInput)
	}
	if b.mtd.IsClientStreaming() {
		_ = b.makeClientStreamingRequest(&ctx, streamInput)
	}
	if b.mtd.IsServerStreaming() {
		_ = b.makeServerStreamingRequest(&ctx, input)
	}

	// TODO: handle response?
	_, _ = b.stub.InvokeRpc(ctx, b.mtd, input)
	return err
}

func (b *Requester) makeClientStreamingRequest(ctx *context.Context, input *[]*dynamic.Message) error {
	str, err := b.stub.InvokeRpcClientStream(*ctx, b.mtd)
	counter := 0
	// TODO: need to handle and propagate errors
	for err == nil {
		streamInput := *input
		inputLen := len(streamInput)
		if input == nil || inputLen == 0 {
			// TODO: need to handle error
			_, _ = str.CloseAndReceive()
			break
		}

		if counter == inputLen {
			// TODO: need to handle error
			_, _ = str.CloseAndReceive()
			break
		}

		payload := streamInput[counter]

		var wait <-chan time.Time
		if b.config.streamInterval > 0 {
			wait = time.Tick(b.config.streamInterval)
			<-wait
		}

		err = str.SendMsg(payload)
		if err == io.EOF {
			// We get EOF on send if the server says "go away"
			// We have to use CloseAndReceive to get the actual code
			// TODO: need to handle error
			_, _ = str.CloseAndReceive()
			break
		}
		counter++
	}
	return nil
}

func (b *Requester) makeServerStreamingRequest(ctx *context.Context, input *dynamic.Message) error {
	str, err := b.stub.InvokeRpcServerStream(*ctx, b.mtd, input)
	// TODO: need to handle and propagate errors
	for err == nil {
		_, err := str.RecvMsg()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
	return nil
}

func (b *Requester) makeBidiRequest(ctx *context.Context, input *[]*dynamic.Message) error {
	str, err := b.stub.InvokeRpcBidiStream(*ctx, b.mtd)
	if err != nil {
		return err
	}

	counter := 0

	streamInput := *input
	inputLen := len(streamInput)

	recvDone := make(chan bool)

	if input == nil || inputLen == 0 {
		// TODO: need to handle error
		_ = str.CloseSend()
		return nil
	}

	go func() {
		for {
			_, err := str.RecvMsg()

			if err != nil {
				close(recvDone)
				break
			}
		}
	}()

	// TODO: need to handle and propagate errors
	for err == nil {
		if counter == inputLen {
			// TODO: need to handle error
			_ = str.CloseSend()
			break
		}

		payload := streamInput[counter]

		var wait <-chan time.Time
		if b.config.streamInterval > 0 {
			wait = time.Tick(b.config.streamInterval)
			<-wait
		}

		err = str.SendMsg(payload)
		counter++
	}

	if err == nil {
		<-recvDone
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
