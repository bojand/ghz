package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	dataProvider     DataProviderFunc
	metadataProvider MetadataProviderFunc
	streamRecv       StreamRecvMsgInterceptFunc
}

func (w *Worker) runWorker() error {
	var err error
	g := new(errgroup.Group)

	for {
		select {
		case <-w.stopCh:
			if w.config.async {
				return g.Wait()
			}

			return err
		case tv := <-w.ticks:
			if w.config.async {
				g.Go(func() error {
					return w.makeRequest(tv)
				})
			} else {
				rErr := w.makeRequest(tv)
				err = multierr.Append(err, rErr)
			}
		}
	}
}

// Stop stops the worker. It has to be started with Run() again.
func (w *Worker) Stop() {
	if !w.active {
		return
	}

	w.active = false
	w.stopCh <- true
}

func (w *Worker) makeRequest(tv TickValue) error {
	reqNum := int64(tv.reqNumber)

	ctd := newCallData(w.mtd, w.config.funcs, w.workerID, reqNum)

	inputs, err := w.dataProvider(ctd)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no data provided for request")
	}

	reqMD, err := w.metadataProvider(ctd)
	if err != nil {
		return err
	}

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

	unaryInput := inputs[0]
	var msgProvider StreamMessageProviderFunc
	if w.mtd.IsClientStreaming() {
		if w.config.streamDynamicMessages {
			mp, err := newDynamicMessageProvider(w.mtd, w.config.data, w.config.streamCallCount)
			if err != nil {
				return err
			}

			msgProvider = mp.GetStreamMessage
		} else {
			mp, err := newStaticMessageProvider(w.config.streamCallCount, inputs)
			if err != nil {
				return err
			}

			msgProvider = mp.GetStreamMessage
		}
	}

	// RPC errors are handled via stats handler
	if w.mtd.IsClientStreaming() && w.mtd.IsServerStreaming() {
		_ = w.makeBidiRequest(&ctx, ctd, msgProvider)
	} else if w.mtd.IsClientStreaming() {
		_ = w.makeClientStreamingRequest(&ctx, ctd, msgProvider)
	} else if w.mtd.IsServerStreaming() {
		_ = w.makeServerStreamingRequest(&ctx, unaryInput)
	} else {
		_ = w.makeUnaryRequest(&ctx, reqMD, unaryInput)
	}

	return err
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

func (w *Worker) makeClientStreamingRequest(ctx *context.Context,
	ctd *CallData, messageProvider StreamMessageProviderFunc) error {
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

	closeStream := func() {
		res, closeErr := str.CloseAndReceive()

		if w.config.hasLog {
			w.config.log.Debugw("Close and receive", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"response", res, "error", closeErr)
		}
	}

	performSend := func(payload *dynamic.Message) bool {
		err = str.SendMsg(payload)

		if w.config.hasLog {
			w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"payload", payload, "error", err)
		}

		if err == io.EOF {
			return true
		}

		return false
	}

	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			<-sct.C
			cancel <- struct{}{}
		}()
	}

	done := false
	for err == nil && !done {
		payload, err := messageProvider(ctd)
		if errors.Is(err, ErrEndStream) {
			closeStream()
			break
		}

		if end := performSend(payload); end {
			closeStream()
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
		} else if w.config.streamCallDuration > 0 && len(cancel) > 0 {
			<-cancel
			closeStream()
			done = true
		}
	}

	close(cancel)

	return nil
}

func (w *Worker) makeServerStreamingRequest(ctx *context.Context, input *dynamic.Message) error {
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}

	callCtx, callCancel := context.WithCancel(*ctx)
	defer callCancel()

	str, err := w.stub.InvokeRpcServerStream(callCtx, w.mtd, input, callOptions...)

	if err != nil && w.config.hasLog {
		w.config.log.Errorw("Invoke Server Streaming RPC call error: "+err.Error(), "workerID", w.workerID,
			"call type", "server-streaming",
			"call", w.mtd.GetFullyQualifiedName(),
			"input", input, "error", err)
	}

	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			<-sct.C
			cancel <- struct{}{}
		}()
	}

	interceptCanceled := false
	counter := uint(0)
	for err == nil {
		res, err := str.RecvMsg()

		if w.config.hasLog {
			w.config.log.Debugw("Receive message", "workerID", w.workerID, "call type", "server-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"response", res, "error", err)
		}

		// with any of the cancellation operations we can't just bail
		// we have to drain the messages until the server gets the cancel and ends their side of the stream

		if w.streamRecv != nil {
			if converted, ok := res.(*dynamic.Message); ok {
				err = w.streamRecv(converted, err)
				if errors.Is(err, ErrEndStream) && !interceptCanceled {
					interceptCanceled = true
					err = nil

					callCancel()
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				err = nil
			}

			break
		}

		counter++

		if w.config.streamCallDuration > 0 && len(cancel) > 0 {
			<-cancel
			callCancel()
		} else if w.config.streamCallCount > 0 && counter >= w.config.streamCallCount {
			callCancel()
		}
	}

	close(cancel)

	return err
}

func (w *Worker) makeBidiRequest(ctx *context.Context,
	ctd *CallData, messageProvider StreamMessageProviderFunc) error {
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
	indexCounter := 0
	recvDone := make(chan bool)
	sendDone := make(chan bool)

	closeStream := func() {
		closeErr := str.CloseSend()

		if w.config.hasLog {
			w.config.log.Debugw("Close send", "workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(), "error", closeErr)
		}
	}

	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			<-sct.C
			cancel <- struct{}{}
		}()
	}

	go func() {
		interceptCanceled := false

		for {
			res, err := str.RecvMsg()

			if w.config.hasLog {
				w.config.log.Debugw("Receive message", "workerID", w.workerID, "call type", "bidi",
					"call", w.mtd.GetFullyQualifiedName(),
					"response", res, "error", err)
			}

			if w.streamRecv != nil {
				if converted, ok := res.(*dynamic.Message); ok {
					err = w.streamRecv(converted, err)
					if errors.Is(err, ErrEndStream) && !interceptCanceled {
						interceptCanceled = true
						cancel <- struct{}{}
						err = nil
					}
				}
			}

			if err != nil {
				close(recvDone)
				break
			}
		}
	}()

	go func() {
		done := false

		for err == nil && !done {
			payload, err := messageProvider(ctd)
			if errors.Is(err, ErrEndStream) {
				closeStream()
				break
			}

			err = str.SendMsg(payload)

			counter++
			indexCounter++

			if w.config.hasLog {
				w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "bidi",
					"call", w.mtd.GetFullyQualifiedName(),
					"payload", payload, "error", err)
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
			} else if len(cancel) > 0 {
				fmt.Println("received cancel in sender")
				<-cancel
				closeStream()
				done = true
			}
		}

		close(sendDone)
	}()

	if err == nil {
		recv, send := false, false
		for !recv || !send {
			select {
			case <-recvDone:
				recv = true
			case <-sendDone:
				send = true
			}
		}
	}

	return nil
}
