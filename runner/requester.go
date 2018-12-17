package runner

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
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
}

func newRequester(mtd *desc.MethodDescriptor, c *RunConfig) (*Requester, error) {
	md := mtd.GetInputType()
	payloadMessage := dynamic.NewMessage(md)
	if payloadMessage == nil {
		return nil, fmt.Errorf("No input type of method: %s", mtd.GetName())
	}

	var qpsTick time.Duration
	if c.qps > 0 {
		qpsTick = time.Duration(1e6/(c.qps)) * time.Microsecond
	}

	reqr := &Requester{
		config:     c,
		mtd:        mtd,
		qpsTick:    qpsTick,
		stopReason: ReasonNormalEnd}

	return reqr, nil
}

// Run makes all the requests and returns a report of results
// It blocks until all work is done.
func (b *Requester) Run() (*Report, error) {
	b.results = make(chan *callResult, min(b.config.c*1000, maxResult))
	b.stopCh = make(chan bool, b.config.c)
	b.start = time.Now()

	cc, err := b.connect()
	if err != nil {
		return nil, err
	}

	b.cc = cc
	defer cc.Close()

	b.stub = grpcdynamic.NewStub(cc)

	b.reporter = newReporter(b.results, b.config)

	go func() {
		b.reporter.Run()
	}()

	b.runWorkers()

	report := b.Finish()

	return report, nil
}

// Stop stops the test
func (b *Requester) Stop(reason StopReason) {
	// Send stop signal so that workers can stop gracefully.
	for i := 0; i < b.config.c; i++ {
		b.stopCh <- true
	}

	b.stopReason = reason
}

// Finish finishes the test run
func (b *Requester) Finish() *Report {
	close(b.results)
	total := time.Now().Sub(b.start)

	// Wait until the reporter is done.
	<-b.reporter.done

	return b.reporter.Finalize(b.stopReason, total)
}

func (b *Requester) connect() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	credOptions, err := createClientCredOption(b.config)
	if err != nil {
		return nil, err
	}

	opts = append(opts, credOptions)

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

	opts = append(opts, grpc.WithStatsHandler(&statsHandler{b.results}))

	// create client connection
	return grpc.DialContext(ctx, b.config.host, opts...)
}

func (b *Requester) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.config.c)

	nReqPerWorker := b.config.n / b.config.c

	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.config.c; i++ {
		go func() {
			defer wg.Done()

			b.runWorker(nReqPerWorker)
		}()
	}
	wg.Wait()
}

func (b *Requester) runWorker(n int) {
	var throttle <-chan time.Time
	if b.config.qps > 0 {
		throttle = time.Tick(b.qpsTick)
	}

	for i := 0; i < n; i++ {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-b.stopCh:
			return
		default:
			if b.config.qps > 0 {
				<-throttle
			}

			err := b.makeRequest()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, _ = context.WithTimeout(ctx, b.config.timeout)

	// include the metadata
	if reqMD != nil {
		ctx = metadata.NewOutgoingContext(ctx, *reqMD)
	}

	if b.mtd.IsClientStreaming() && b.mtd.IsServerStreaming() {
		b.makeBidiRequest(&ctx, streamInput)
	} else if b.mtd.IsClientStreaming() {
		b.makeClientStreamingRequest(&ctx, streamInput)
	} else if b.mtd.IsServerStreaming() {
		b.makeServerStreamingRequest(&ctx, input)
	} else {
		b.stub.InvokeRpc(ctx, b.mtd, input)
	}

	return nil
}

func (b *Requester) makeClientStreamingRequest(ctx *context.Context, input *[]*dynamic.Message) {
	str, err := b.stub.InvokeRpcClientStream(*ctx, b.mtd)
	counter := 0
	for err == nil {
		streamInput := *input
		inputLen := len(streamInput)
		if input == nil || inputLen == 0 {
			str.CloseAndReceive()
			break
		}

		if counter == inputLen {
			str.CloseAndReceive()
			break
		}

		payload := streamInput[counter]
		err = str.SendMsg(payload)
		if err == io.EOF {
			// We get EOF on send if the server says "go away"
			// We have to use CloseAndReceive to get the actual code
			str.CloseAndReceive()
			break
		}
		counter++
	}
}

func (b *Requester) makeServerStreamingRequest(ctx *context.Context, input *dynamic.Message) {
	str, err := b.stub.InvokeRpcServerStream(*ctx, b.mtd, input)
	for err == nil {
		_, err := str.RecvMsg()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
}

func (b *Requester) makeBidiRequest(ctx *context.Context, input *[]*dynamic.Message) {
	str, err := b.stub.InvokeRpcBidiStream(*ctx, b.mtd)
	counter := 0
	for err == nil {
		streamInput := *input
		inputLen := len(streamInput)
		if input == nil || inputLen == 0 {
			str.CloseSend()
			break
		}

		if counter == inputLen {
			str.CloseSend()
			break
		}

		payload := streamInput[counter]
		err = str.SendMsg(payload)
		counter++
	}

	for err == nil {
		_, err := str.RecvMsg()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}
}

func createClientCredOption(config *RunConfig) (grpc.DialOption, error) {
	if config.insecure {
		credOptions := grpc.WithInsecure()
		return credOptions, nil
	}

	if config.cert != "" {
		creds, err := credentials.NewClientTLSFromFile(config.cert, config.cname)
		if err != nil {
			return nil, err
		}
		credOptions := grpc.WithTransportCredentials(creds)
		return credOptions, nil
	}

	credOptions := grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	return credOptions, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
