package runner

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

// TickValue is the tick value
type TickValue struct {
	instant   time.Time
	reqNumber uint64
}

// Worker is used for doing a single stream of requests in parallel
type Worker struct {
	stub grpcdynamic.Stub
	mtd  *desc.MethodDescriptor

	config   *RunConfig
	workerID string
	active   bool
	stopCh   chan bool
	ticks    <-chan TickValue

	// cached messages only for binary
	cachedMessages []*dynamic.Message

	// non-binary json optimization
	arrayJSONData []string
}

func (w *Worker) runWorker() error {
	var err error
	g := new(errgroup.Group)

	now := time.Now()

	for {
		select {
		case <-w.stopCh:
			if w.config.async {
				return g.Wait()
			}

			fmt.Println("run worker done", time.Since(now).String())
			return err

		case tv := <-w.ticks:
			if w.config.async {
				g.Go(func() error {
					return w.makeRequest(tv)
				})
			} else {
				fmt.Println("tick make request", time.Since(now).String())
				rErr := w.makeRequest(tv)
				fmt.Println("tick made request. appending", time.Since(now).String())
				err = multierr.Append(err, rErr)
				fmt.Println("tick made request appended", time.Since(now).String())
			}
		}
	}
}

// Stop stops the worker. It has to be started with Run() again.
func (w *Worker) Stop() {
	now := time.Now()
	fmt.Println("Worker Stop()", w.active)
	if !w.active {
		return
	}

	fmt.Println("Worker Stopping", time.Since(now).String())
	w.active = false
	w.stopCh <- true
	fmt.Println("Worker Stoppped", time.Since(now).String())
}

func (w *Worker) makeRequest(tv TickValue) error {
	now := time.Now()
	fmt.Println(w.workerID, "makeRequest")
	reqNum := int64(tv.reqNumber)

	ctd := newCallData(w.mtd, w.config.funcs, w.workerID, reqNum)

	fmt.Println(w.workerID, "made call data", time.Since(now).String())

	var inputs []*dynamic.Message
	var err error

	// try the optimized path for JSON data for non client-streaming
	if !w.config.binary && !w.mtd.IsClientStreaming() && len(w.arrayJSONData) > 0 {
		indx := int(reqNum % int64(len(w.arrayJSONData))) // we want to start from inputs[0] so dec reqNum
		if inputs, err = w.getMessages(ctd, []byte(w.arrayJSONData[indx])); err != nil {
			return err
		}
	} else {
		if inputs, err = w.getMessages(ctd, w.config.data); err != nil {
			return err
		}
	}

	fmt.Println(w.workerID, "made call inputs", time.Since(now).String())

	mdMap, err := ctd.executeMetadata(string(w.config.metadata))
	if err != nil {
		return err
	}

	var reqMD *metadata.MD
	if len(mdMap) > 0 {
		md := metadata.New(mdMap)
		reqMD = &md
	} else {
		reqMD = &metadata.MD{}
	}

	fmt.Println(w.workerID, "made call metadata", time.Since(now).String())

	if w.config.enableCompression {
		reqMD.Append("grpc-accept-encoding", gzip.Name)
	}

	ctx := context.Background()
	var cancel context.CancelFunc

	if w.config.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, w.config.timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// include the metadata
	if reqMD != nil {
		ctx = metadata.NewOutgoingContext(ctx, *reqMD)
	}

	var callType string
	if w.config.hasLog {
		callType = "unary"
		if w.mtd.IsClientStreaming() && w.mtd.IsServerStreaming() {
			callType = "bidi"
		} else if w.mtd.IsServerStreaming() {
			callType = "server-streaming"
		} else if w.mtd.IsClientStreaming() {
			callType = "client-streaming"
		}

		w.config.log.Debugw("Making request", "workerID", w.workerID,
			"call type", callType, "call", w.mtd.GetFullyQualifiedName(),
			"input", inputs, "metadata", reqMD)
	}

	inputsLen := len(inputs)
	if inputsLen == 0 {
		return fmt.Errorf("no data provided for request")
	}
	inputIdx := int(reqNum % int64(inputsLen)) // we want to start from inputs[0] so dec reqNum
	unaryInput := inputs[inputIdx]

	start := time.Now()
	fmt.Println("starting client streaming", time.Since(now).String())
	// RPC errors are handled via stats handler
	if w.mtd.IsClientStreaming() && w.mtd.IsServerStreaming() {
		_ = w.makeBidiRequest(&ctx, inputs)
	} else if w.mtd.IsClientStreaming() {
		_ = w.makeClientStreamingRequest(&ctx, inputs)
	} else if w.mtd.IsServerStreaming() {
		_ = w.makeServerStreamingRequest(&ctx, unaryInput)
	} else {
		_ = w.makeUnaryRequest(&ctx, reqMD, unaryInput)
	}

	fmt.Println("make request done", time.Since(now).String(), time.Since(start))

	return err
}

func (w *Worker) getMessages(ctd *CallData, inputData []byte) ([]*dynamic.Message, error) {
	var inputs []*dynamic.Message

	if w.cachedMessages != nil {
		return w.cachedMessages, nil
	}

	if !w.config.binary {
		now := time.Now()
		data, err := ctd.executeData(string(inputData))
		if err != nil {
			return nil, err
		}
		fmt.Println("done execute", time.Since(now))
		now = time.Now()
		inputs, err = createPayloadsFromJSON(string(data), w.mtd)
		if err != nil {
			return nil, err
		}
		fmt.Println("done createPayloadsFromJSON", time.Since(now))
		// JSON messages are not cached due to templating
	} else {
		var err error
		if w.config.dataFunc != nil {
			inputData = w.config.dataFunc(w.mtd, ctd)
		}
		inputs, err = createPayloadsFromBin(inputData, w.mtd)
		if err != nil {
			return nil, err
		}
		// We only cache in case we don't dynamically change the binary message
		if w.config.dataFunc == nil {
			w.cachedMessages = inputs
		}
	}

	return inputs, nil
}

func (w *Worker) makeUnaryRequest(ctx *context.Context, reqMD *metadata.MD, input *dynamic.Message) error {
	var res proto.Message
	var resErr error
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}

	res, resErr = w.stub.InvokeRpc(*ctx, w.mtd, input, callOptions...)

	if w.config.hasLog {
		w.config.log.Debugw("Received response", "workerID", w.workerID, "call type", "unary",
			"call", w.mtd.GetFullyQualifiedName(),
			"input", input, "metadata", reqMD,
			"response", res, "error", resErr)
	}

	return resErr
}

func (w *Worker) makeClientStreamingRequest(ctx *context.Context, input []*dynamic.Message) error {
	fmt.Println("makeClientStreamingRequest()")
	var str *grpcdynamic.ClientStream
	var err error
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	str, err = w.stub.InvokeRpcClientStream(*ctx, w.mtd, callOptions...)

	if err != nil && w.config.hasLog {
		w.config.log.Errorw("Invoke Client Streaming RPC call error: "+err.Error(), "workerID", w.workerID,
			"call type", "client-streaming",
			"call", w.mtd.GetFullyQualifiedName(), "error", err)
	}

	counter := 0

	closeStream := func() {
		res, closeErr := str.CloseAndReceive()

		if w.config.hasLog {
			w.config.log.Debugw("Close and receive", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"response", res, "error", closeErr)
		}
	}

	performSend := func() bool {
		// fmt.Println("performSend", counter)
		inputLen := len(input)
		if input == nil || inputLen == 0 {
			return true
		}

		if counter == inputLen {
			return true
		}

		payload := input[counter]

		err = str.SendMsg(payload)

		if w.config.hasLog {
			w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"payload", payload, "error", err)
		}

		if err == io.EOF {
			return true
		}

		counter++

		return false
	}

	cancel := make(chan struct{}, 1)
	fmt.Println(w.config.streamClose.String())
	if w.config.streamClose > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamClose)
			<-sct.C
			cancel <- struct{}{}
		}()
	}

	done := false
	start := time.Now()
	for err == nil && !done {

		if end := performSend(); end {
			closeStream()
			done = true
			break
		}

		if w.config.streamInterval > 0 {
			wait := time.NewTimer(w.config.streamInterval)
			select {
			case <-wait.C:
				break
			case <-cancel:
				closeStream()
				done = true
			}
		} else if w.config.streamClose > 0 && len(cancel) > 0 {
			<-cancel
			closeStream()
			done = true
		}

		if done {
			break
		}
	}

	// fmt.Println(len(cancel), time.Since(start).String(), done)

	close(cancel)

	fmt.Println("returning", len(cancel), time.Since(start).String(), done, "counter:", counter)

	return nil
}

func (w *Worker) makeServerStreamingRequest(ctx *context.Context, input *dynamic.Message) error {
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	str, err := w.stub.InvokeRpcServerStream(*ctx, w.mtd, input, callOptions...)

	if err != nil && w.config.hasLog {
		w.config.log.Errorw("Invoke Server Streaming RPC call error: "+err.Error(), "workerID", w.workerID,
			"call type", "server-streaming",
			"call", w.mtd.GetFullyQualifiedName(),
			"input", input, "error", err)
	}

	for err == nil {
		res, err := str.RecvMsg()

		if w.config.hasLog {
			w.config.log.Debugw("Receive message", "workerID", w.workerID, "call type", "server-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"response", res, "error", err)
		}

		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
	}

	return err
}

func (w *Worker) makeBidiRequest(ctx *context.Context, input []*dynamic.Message) error {
	var str *grpcdynamic.BidiStream
	var err error
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	str, err = w.stub.InvokeRpcBidiStream(*ctx, w.mtd, callOptions...)

	if err != nil {
		if w.config.hasLog {
			w.config.log.Errorw("Invoke Bidi RPC call error: "+err.Error(),
				"workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(), "error", err)
		}

		return err
	}

	counter := 0

	inputLen := len(input)

	recvDone := make(chan bool)

	if input == nil || inputLen == 0 {
		closeErr := str.CloseSend()

		if w.config.hasLog {
			w.config.log.Debugw("Close send", "workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(), "error", closeErr)
		}

		return nil
	}

	go func() {
		for {
			res, err := str.RecvMsg()

			if w.config.hasLog {
				w.config.log.Debugw("Receive message", "workerID", w.workerID, "call type", "bidi",
					"call", w.mtd.GetFullyQualifiedName(),
					"response", res, "error", err)
			}

			if err != nil {
				close(recvDone)
				break
			}
		}
	}()

	closeStream := func() {
		closeErr := str.CloseSend()

		if w.config.hasLog {
			w.config.log.Debugw("Close send", "workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(), "error", closeErr)
		}
	}

	var finished uint32
	if w.config.streamClose > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamClose)
			<-sct.C
			atomic.AddUint32(&finished, 1)
		}()
	}

	for err == nil {

		if counter == inputLen {
			closeStream()
			break
		}

		payload := input[counter]

		// we need to check before and after stream interval
		toClose := atomic.LoadUint32(&finished)
		if toClose > 0 {
			closeStream()
			break
		}

		var wait <-chan time.Time
		if w.config.streamInterval > 0 {
			wait = time.Tick(w.config.streamInterval)
			<-wait
		}

		toClose = atomic.LoadUint32(&finished)
		if toClose > 0 {
			closeStream()
			break
		}

		err = str.SendMsg(payload)
		counter++

		if w.config.hasLog {
			w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(),
				"payload", payload, "error", err)
		}
	}

	if err == nil {
		<-recvDone
	}

	return nil
}
