package runner

import (
	"context"
	"encoding/json"
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
	msgProvider      StreamMessageProviderFunc

	streamRecv                    StreamRecvMsgInterceptFunc
	streamInterceptorProviderFunc StreamInterceptorProviderFunc
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

	ctd := newCallData(w.mtd, w.workerID, reqNum, !w.config.disableTemplateFuncs, !w.config.disableTemplateData, w.config.funcs)

	var streamInterceptor StreamInterceptor
	if w.mtd.IsClientStreaming() || w.mtd.IsServerStreaming() {
		if w.streamInterceptorProviderFunc != nil {
			streamInterceptor = w.streamInterceptorProviderFunc(ctd)
		}
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

	inputs, err := w.dataProvider(ctd)
	if err != nil {
		return err
	}

	var msgProvider StreamMessageProviderFunc
	if w.msgProvider != nil {
		msgProvider = w.msgProvider
	} else if streamInterceptor != nil {
		msgProvider = streamInterceptor.Send
	} else if w.mtd.IsClientStreaming() {
		if w.config.streamDynamicMessages {
			mp, err := newDynamicMessageProvider(w.mtd, w.config.data, w.config.streamCallCount, !w.config.disableTemplateFuncs, !w.config.disableTemplateData)
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

	if len(inputs) == 0 && msgProvider == nil {
		return fmt.Errorf("no data provided for request")
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

	// RPC errors are handled via stats handler
	if w.mtd.IsClientStreaming() && w.mtd.IsServerStreaming() {
		_ = w.makeBidiRequest(&ctx, ctd, msgProvider, streamInterceptor)
	} else if w.mtd.IsClientStreaming() {
		_ = w.makeClientStreamingRequest(&ctx, ctd, msgProvider)
	} else if w.mtd.IsServerStreaming() {
		_ = w.makeServerStreamingRequest(&ctx, inputs[0], streamInterceptor)
	} else {
		_ = w.makeUnaryRequest(&ctx, reqMD, inputs[0])
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
		inputData, _ := input.MarshalJSON()
		resData, _ := json.Marshal(res)

		w.config.log.Debugw("Received response", "workerID", w.workerID, "call type", "unary",
			"call", w.mtd.GetFullyQualifiedName(),
			"input", string(inputData), "metadata", reqMD,
			"response", string(resData), "error", resErr)
	}

	return resErr
}

func (w *Worker) makeClientStreamingRequest(ctx *context.Context,
	ctd *CallData, messageProvider StreamMessageProviderFunc) error {
	var str *grpcdynamic.ClientStream
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	str, err := w.stub.InvokeRpcClientStream(*ctx, w.mtd, callOptions...)
	if err != nil {
		if w.config.hasLog {
			w.config.log.Errorw("Invoke Client Streaming RPC call error: "+err.Error(), "workerID", w.workerID,
				"call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(), "error", err)
		}

		return err
	}

	closeStream := func() {
		res, closeErr := str.CloseAndReceive()

		if w.config.hasLog {
			w.config.log.Debugw("Close and receive", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"response", res, "error", closeErr)
		}
	}

	performSend := func(payload *dynamic.Message) (bool, error) {
		err := str.SendMsg(payload)

		if w.config.hasLog {
			w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "client-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"payload", payload, "error", err)
		}

		if err == io.EOF {
			return true, nil
		}

		return false, err
	}

	doneCh := make(chan struct{})
	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			select {
			case <-sct.C:
				cancel <- struct{}{}
				return
			case <-doneCh:
				if !sct.Stop() {
					<-sct.C
				}
				return
			}
		}()
	}

	done := false
	counter := uint(0)
	end := false
	for !done && len(cancel) == 0 {
		// default message provider checks counter
		// but we also need to keep our own counts
		// in case of custom client providers

		var payload *dynamic.Message
		payload, err = messageProvider(ctd)

		isLast := false
		if errors.Is(err, ErrLastMessage) {
			isLast = true
			err = nil
		}

		if err != nil {
			if errors.Is(err, ErrEndStream) {
				err = nil
			}
			break
		}

		end, err = performSend(payload)
		if end || err != nil || isLast || len(cancel) > 0 {
			break
		}

		counter++

		if w.config.streamCallCount > 0 && counter >= w.config.streamCallCount {
			break
		}

		if w.config.streamInterval > 0 {
			wait := time.NewTimer(w.config.streamInterval)
			select {
			case <-wait.C:
				break
			case <-cancel:
				if !wait.Stop() {
					<-wait.C
				}
				done = true
				break
			}
		}
	}

	for len(cancel) > 0 {
		<-cancel
	}

	closeStream()

	close(doneCh)
	close(cancel)

	return nil
}

func (w *Worker) makeServerStreamingRequest(ctx *context.Context, input *dynamic.Message, streamInterceptor StreamInterceptor) error {
	var callOptions = []grpc.CallOption{}
	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}

	callCtx, callCancel := context.WithCancel(*ctx)
	defer callCancel()

	str, err := w.stub.InvokeRpcServerStream(callCtx, w.mtd, input, callOptions...)

	if err != nil {
		if w.config.hasLog {
			w.config.log.Errorw("Invoke Server Streaming RPC call error: "+err.Error(), "workerID", w.workerID,
				"call type", "server-streaming",
				"call", w.mtd.GetFullyQualifiedName(),
				"input", input, "error", err)
		}

		return err
	}

	doneCh := make(chan struct{})
	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			select {
			case <-sct.C:
				cancel <- struct{}{}
				return
			case <-doneCh:
				if !sct.Stop() {
					<-sct.C
				}
				return
			}
		}()
	}

	interceptCanceled := false
	counter := uint(0)
	for err == nil {
		// we should check before receiving a message too
		if w.config.streamCallDuration > 0 && len(cancel) > 0 {
			<-cancel
			callCancel()
			break
		}

		var res proto.Message
		res, err = str.RecvMsg()

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

		if streamInterceptor != nil {
			if converted, ok := res.(*dynamic.Message); ok {
				err = streamInterceptor.Recv(converted, err)
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

		if w.config.streamCallCount > 0 && counter >= w.config.streamCallCount {
			callCancel()
		}

		if w.config.streamCallDuration > 0 && len(cancel) > 0 {
			<-cancel
			callCancel()
		}
	}

	close(doneCh)
	close(cancel)

	return err
}

func (w *Worker) makeBidiRequest(ctx *context.Context,
	ctd *CallData, messageProvider StreamMessageProviderFunc, streamInterceptor StreamInterceptor) error {

	var callOptions = []grpc.CallOption{}

	if w.config.enableCompression {
		callOptions = append(callOptions, grpc.UseCompressor(gzip.Name))
	}
	str, err := w.stub.InvokeRpcBidiStream(*ctx, w.mtd, callOptions...)

	if err != nil {
		if w.config.hasLog {
			w.config.log.Errorw("Invoke Bidi RPC call error: "+err.Error(),
				"workerID", w.workerID, "call type", "bidi",
				"call", w.mtd.GetFullyQualifiedName(), "error", err)
		}

		return err
	}

	counter := uint(0)
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

	doneCh := make(chan struct{})
	cancel := make(chan struct{}, 1)
	if w.config.streamCallDuration > 0 {
		go func() {
			sct := time.NewTimer(w.config.streamCallDuration)
			select {
			case <-sct.C:
				cancel <- struct{}{}
				return
			case <-doneCh:
				if !sct.Stop() {
					<-sct.C
				}
				return
			}
		}()
	}

	var recvErr error

	go func() {
		interceptCanceled := false

		for recvErr == nil {
			var res proto.Message
			res, recvErr = str.RecvMsg()

			if w.config.hasLog {
				w.config.log.Debugw("Receive message", "workerID", w.workerID, "call type", "bidi",
					"call", w.mtd.GetFullyQualifiedName(),
					"response", res, "error", recvErr)
			}

			if w.streamRecv != nil {
				if converted, ok := res.(*dynamic.Message); ok {
					iErr := w.streamRecv(converted, recvErr)
					if errors.Is(iErr, ErrEndStream) && !interceptCanceled {
						interceptCanceled = true
						if len(cancel) == 0 {
							cancel <- struct{}{}
						}
						recvErr = nil
					}
				}
			}

			if streamInterceptor != nil {
				if converted, ok := res.(*dynamic.Message); ok {
					iErr := streamInterceptor.Recv(converted, recvErr)
					if errors.Is(iErr, ErrEndStream) && !interceptCanceled {
						interceptCanceled = true
						if len(cancel) == 0 {
							cancel <- struct{}{}
						}
						recvErr = nil
					}
				}
			}

			if recvErr != nil {
				close(recvDone)
				break
			}
		}
	}()

	go func() {
		done := false

		for err == nil && !done {

			// check at start before send too
			if len(cancel) > 0 {
				<-cancel
				closeStream()
				break
			}

			// default message provider checks counter
			// but we also need to keep our own counts
			// in case of custom client providers

			var payload *dynamic.Message
			payload, err = messageProvider(ctd)

			isLast := false
			if errors.Is(err, ErrLastMessage) {
				isLast = true
				err = nil
			}

			if err != nil {
				if errors.Is(err, ErrEndStream) {
					err = nil
				}

				closeStream()
				break
			}

			err = str.SendMsg(payload)
			if err != nil {
				if err == io.EOF {
					err = nil
				}

				break
			}

			if w.config.hasLog {
				w.config.log.Debugw("Send message", "workerID", w.workerID, "call type", "bidi",
					"call", w.mtd.GetFullyQualifiedName(),
					"payload", payload, "error", err)
			}

			if isLast {
				closeStream()
				break
			}

			counter++
			indexCounter++

			if w.config.streamCallCount > 0 && counter >= w.config.streamCallCount {
				closeStream()
				break
			}

			if len(cancel) > 0 {
				<-cancel
				closeStream()
				break
			}

			if w.config.streamInterval > 0 {
				wait := time.NewTimer(w.config.streamInterval)
				select {
				case <-wait.C:
					break
				case <-cancel:
					if !wait.Stop() {
						<-wait.C
					}
					closeStream()
					done = true
					break
				}
			}
		}

		close(sendDone)
	}()

	_, _ = <-recvDone, <-sendDone

	for len(cancel) > 0 {
		<-cancel
	}

	close(doneCh)
	close(cancel)

	if err == nil && recvErr != nil {
		err = recvErr
	}

	return err
}
