package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc/metadata"
)

// TODO move to own pacakge?
// TODO add tests
// TODO expose public API utilizing only proto API and not dynamic

// ErrEndStream is a signal from message providers that worker should close the stream
// It should not be used for erronous states
var ErrEndStream = errors.New("ending stream")

// ErrLastMessage is a signal from message providers that the returned payload is the last one of the stream
// This is optional but encouraged for optimized performance
// Message payload returned along with this error must be valid and may not be nil
var ErrLastMessage = errors.New("last message")

// DataProviderFunc is the interface for providing data for calls
// For unary and server streaming calls it should return an array with a single element
// For client and bidi streaming calls it should return an array of messages to be used
type DataProviderFunc func(*CallData) ([]*dynamic.Message, error)

// MetadataProviderFunc is the interface for providing metadadata for calls
type MetadataProviderFunc func(*CallData) (*metadata.MD, error)

// StreamMessageProviderFunc is the interface for providing a message for every message send in the course of a streaming call
type StreamMessageProviderFunc func(*CallData) (*dynamic.Message, error)

// StreamRecvMsgInterceptFunc is an interface for function invoked when we receive a stream message
// Clients can return ErrEndStream to end the call early
type StreamRecvMsgInterceptFunc func(*dynamic.Message, error) error

// StreamInterceptorProviderFunc is an interface for a function invoked to generate a stream interceptor
type StreamInterceptorProviderFunc func(*CallData) StreamInterceptor

// StreamInterceptor is an interface for sending and receiving stream messages.
// The interceptor can keep shared state for the send and receive calls.
type StreamInterceptor interface {
	Recv(*dynamic.Message, error) error
	Send(*CallData) (*dynamic.Message, error)
}

type dataProvider struct {
	binary   bool
	data     []byte
	mtd      *desc.MethodDescriptor
	dataFunc BinaryDataFunc

	arrayJSONData []string
	hasActions    bool

	// cached messages only for binary
	mutex          sync.RWMutex
	cachedMessages []*dynamic.Message
}

type mdProvider struct {
	metadata []byte
	preseed  metadata.MD
}

func newDataProvider(mtd *desc.MethodDescriptor,
	binary bool, dataFunc BinaryDataFunc, data []byte,
	withFuncs, withTemplateData bool, funcs template.FuncMap) (*dataProvider, error) {

	dp := dataProvider{
		binary:         binary,
		dataFunc:       dataFunc,
		mtd:            mtd,
		data:           data,
		cachedMessages: nil,
	}

	// fill in JSON string array data for optimization for non client-streaming
	var err error
	dp.arrayJSONData = nil
	if !dp.binary {
		if strings.IndexRune(string(data), '[') == 0 { // it's an array
			var dat []map[string]interface{}
			if err := json.Unmarshal(data, &dat); err != nil {
				return nil, err
			}

			dp.arrayJSONData = make([]string, len(dat))
			for i, d := range dat {
				var strd []byte
				if strd, err = json.Marshal(d); err != nil {
					return nil, err
				}

				dp.arrayJSONData[i] = string(strd)
			}
		}
	}

	// Test if we can preseed data
	ha := false
	ctd := newCallData(mtd, "", 0, withFuncs, withTemplateData, funcs)

	if withTemplateData {
		if !dp.binary {
			ha, err = ctd.hasAction(string(dp.data))
			if err != nil {
				return nil, err
			}
		}
	}

	dp.hasActions = ha

	if !ha {
		if len(dp.arrayJSONData) > 0 {
			dp.mutex.Lock()
			dp.cachedMessages = make([]*dynamic.Message, len(dp.arrayJSONData))
			dp.mutex.Unlock()
		}

		_, err := dp.getDataForCall(ctd)
		if err != nil {
			return nil, err
		}

	}

	return &dp, nil
}

func (dp *dataProvider) getDataForCall(ctd *CallData) ([]*dynamic.Message, error) {
	var inputs []*dynamic.Message
	var err error

	// try the optimized path for JSON data for non client-streaming
	if !dp.binary && !dp.mtd.IsClientStreaming() && len(dp.arrayJSONData) > 0 {
		indx := int(ctd.RequestNumber % int64(len(dp.arrayJSONData))) // we want to start from inputs[0] so dec reqNum

		if inputs, err = dp.getMessages(ctd, indx, []byte(dp.arrayJSONData[indx])); err != nil {
			return nil, err
		}
	} else if inputs, err = dp.getMessages(ctd, -1, dp.data); err != nil {
		return nil, err
	}

	if !dp.mtd.IsClientStreaming() && len(inputs) > 0 {
		inputIdx := int(ctd.RequestNumber % int64(len(inputs)))
		unaryInput := inputs[inputIdx]

		return []*dynamic.Message{unaryInput}, nil
	}

	return inputs, nil
}

func (dp *dataProvider) getMessages(ctd *CallData, i int, inputData []byte) ([]*dynamic.Message, error) {
	var inputs []*dynamic.Message
	var err error

	dp.mutex.RLock()
	if dp.cachedMessages != nil {
		if (i < 0 && len(dp.cachedMessages) > 0 && dp.cachedMessages[0] != nil) ||
			(i >= 0 && i < len(dp.cachedMessages) && dp.cachedMessages[i] != nil) {
			defer dp.mutex.RUnlock()
			return dp.cachedMessages, nil
		}
	}
	dp.mutex.RUnlock()

	if !dp.binary {
		data := inputData

		if dp.hasActions {
			cdata, err := ctd.ExecuteData(string(inputData))
			if err != nil {
				return nil, err
			}
			data = cdata
		}

		inputs, err = createPayloadsFromJSON(string(data), dp.mtd)
		if err != nil {
			return nil, err
		}

		// only cache JSON data if there are no template actions
		if !dp.hasActions {
			dp.mutex.Lock()
			if i < 0 {
				dp.cachedMessages = inputs
			} else {
				if i >= cap(dp.cachedMessages) {
					nc := make([]*dynamic.Message, len(dp.cachedMessages), (cap(dp.cachedMessages) + i + 1))
					copy(nc, dp.cachedMessages)
					dp.cachedMessages = nc
				}

				if i < len(dp.cachedMessages) {
					dp.cachedMessages[i] = inputs[0]
				} else {
					dp.cachedMessages = append(dp.cachedMessages, inputs...)
				}
			}
			dp.mutex.Unlock()
		}
	} else {
		var err error
		if dp.dataFunc != nil {
			inputData = dp.dataFunc(dp.mtd, ctd)
		}
		inputs, err = createPayloadsFromBin(inputData, dp.mtd)
		if err != nil {
			return nil, err
		}
		// We only cache in case we don't dynamically change the binary message
		if dp.dataFunc == nil {
			dp.mutex.Lock()
			dp.cachedMessages = inputs
			dp.mutex.Unlock()
		}
	}

	return inputs, nil
}

func newMetadataProvider(mtd *desc.MethodDescriptor, mdData []byte, withFuncs, withTemplateData bool, funcs template.FuncMap) (*mdProvider, error) {
	// Test if we can preseed data
	ctd := newCallData(mtd, "", 0, withFuncs, withTemplateData, funcs)
	ha, err := ctd.hasAction(string(mdData))
	if err != nil {
		return nil, err
	}

	var preseed metadata.MD = nil
	if !ha {
		mdMap, err := ctd.executeMetadata(string(mdData))
		if err != nil {
			return nil, err
		}

		if len(mdMap) > 0 {
			preseed = metadata.New(mdMap)
		}
	}

	return &mdProvider{metadata: mdData, preseed: preseed}, nil
}

func (dp *mdProvider) getMetadataForCall(ctd *CallData) (*metadata.MD, error) {
	if dp.preseed != nil {
		return &dp.preseed, nil
	}

	mdMap, err := ctd.executeMetadata(string(dp.metadata))
	if err != nil {
		return nil, err
	}

	if len(mdMap) > 0 {
		md := metadata.New(mdMap)
		return &md, nil
	}

	return &metadata.MD{}, nil
}

// creates a message from a map
// marshal to JSON then use jsonpb to marshal to message
// this way we follow protobuf more closely and allow camelCase properties.
func messageFromMap(input *dynamic.Message, data *map[string]interface{}) error {
	strData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = jsonpb.UnmarshalString(string(strData), input)
	if err != nil {
		return err
	}

	return nil
}

func createPayloadsFromJSON(data string, mtd *desc.MethodDescriptor) ([]*dynamic.Message, error) {
	md := mtd.GetInputType()
	var inputs []*dynamic.Message

	if len(data) > 0 {
		if strings.IndexRune(data, '[') == 0 {
			dataArray := make([]map[string]interface{}, 5)
			err := json.Unmarshal([]byte(data), &dataArray)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling payload. Data: '%v' Error: %v", data, err.Error())
			}

			elems := len(dataArray)
			if elems > 0 {
				inputs = make([]*dynamic.Message, elems)
			}

			for i, elem := range dataArray {
				elemMsg := dynamic.NewMessage(md)
				err := messageFromMap(elemMsg, &elem)
				if err != nil {
					return nil, fmt.Errorf("Error creating message: %v", err.Error())
				}

				inputs[i] = elemMsg
			}
		} else {
			inputs = make([]*dynamic.Message, 1)
			inputs[0] = dynamic.NewMessage(md)
			err := jsonpb.UnmarshalString(data, inputs[0])
			if err != nil {
				return nil, fmt.Errorf("Error creating message from data. Data: '%v' Error: %v", data, err.Error())
			}
		}
	}

	return inputs, nil
}

func createPayloadsFromBinSingleMessage(binData []byte, mtd *desc.MethodDescriptor) ([]*dynamic.Message, error) {
	inputs := make([]*dynamic.Message, 0, 1)
	md := mtd.GetInputType()

	// return empty array if no data
	if len(binData) == 0 {
		return inputs, nil
	}

	// try to unmarshal input as a single message
	singleMessage := dynamic.NewMessage(md)
	err := proto.Unmarshal(binData, singleMessage)
	if err != nil {
		return nil, fmt.Errorf("Error creating message from binary data: %v", err.Error())
	}

	inputs = append(inputs, singleMessage)

	return inputs, nil
}

func createPayloadsFromBinCountDelimited(binData []byte, mtd *desc.MethodDescriptor) ([]*dynamic.Message, error) {
	inputs := make([]*dynamic.Message, 0)
	md := mtd.GetInputType()

	// return empty array if no data
	if len(binData) == 0 {
		return inputs, nil
	}

	// try to unmarshal input as several count-delimited messages
	buffer := proto.NewBuffer(binData)
	for {
		msg := dynamic.NewMessage(md)
		err := buffer.DecodeMessage(msg)

		if err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Error creating message from binary data: %v", err.Error())
		}

		inputs = append(inputs, msg)
	}

	return inputs, nil
}

func createPayloadsFromBin(binData []byte, mtd *desc.MethodDescriptor) ([]*dynamic.Message, error) {
	inputs, err := createPayloadsFromBinCountDelimited(binData, mtd)

	if err == nil && len(inputs) > 0 {
		return inputs, err
	}

	return createPayloadsFromBinSingleMessage(binData, mtd)
}

type dynamicMessageProvider struct {
	mtd           *desc.MethodDescriptor
	data          []byte
	arrayJSONData []string
	arrayLen      uint

	streamCallCount uint
	counter         uint
	indexCounter    uint
}

func newDynamicMessageProvider(mtd *desc.MethodDescriptor, data []byte, streamCallCount uint, withFuncs, withTemplateData bool) (*dynamicMessageProvider, error) {
	mp := dynamicMessageProvider{
		mtd:             mtd,
		data:            data,
		arrayJSONData:   nil,
		streamCallCount: streamCallCount,
	}

	var err error

	if strings.IndexRune(string(data), '[') == 0 { // it's an array
		var dat []map[string]interface{}
		if err := json.Unmarshal(data, &dat); err != nil {
			return nil, err
		}

		mp.arrayJSONData = make([]string, len(dat))
		for i, d := range dat {
			var strd []byte
			if strd, err = json.Marshal(d); err != nil {
				return nil, err
			}

			mp.arrayJSONData[i] = string(strd)
		}
	}

	mp.arrayLen = uint(len(mp.arrayJSONData))

	// Test if we have actions
	ha := false
	ctd := newCallData(mtd, "", 0, withFuncs, withTemplateData, nil)

	if withTemplateData {
		ha, err = ctd.hasAction(string(mp.data))
		if err != nil {
			return nil, err
		}
	}

	if !ha {
		return nil, errors.New("default message provider cannot be used with static data")
	}

	return &mp, nil
}

func (m *dynamicMessageProvider) GetStreamMessage(parentCallData *CallData) (*dynamic.Message, error) {
	isLast := false
	if m.streamCallCount > 0 {
		if m.counter >= m.streamCallCount {
			return nil, ErrEndStream
		} else if m.counter >= m.arrayLen {
			m.indexCounter = 0
		}

		isLast = m.counter == m.streamCallCount-1
	} else if m.counter >= m.arrayLen {
		return nil, ErrEndStream
	} else if m.counter == m.arrayLen-1 {
		isLast = true
	}

	ctd := parentCallData.Regenerate()

	data := string(m.data)

	if m.arrayLen > 0 {
		if m.counter >= m.arrayLen {
			m.indexCounter = 0
		}

		data = m.arrayJSONData[m.indexCounter]
	}

	buf, err := ctd.ExecuteData(data)
	if err != nil {
		return nil, err
	}

	md := m.mtd.GetInputType()
	msg := dynamic.NewMessage(md)
	err = jsonpb.UnmarshalString(string(buf), msg)
	if err != nil {
		return nil, fmt.Errorf("Error creating message from data. Data: '%v' Error: %v", data, err.Error())
	}

	m.counter++
	m.indexCounter++

	if err == nil && isLast {
		err = ErrLastMessage
	}

	return msg, err
}

type staticMessageProvider struct {
	inputs          []*dynamic.Message
	inputLen        uint
	streamCallCount uint
	counter         uint
	indexCounter    uint
}

func newStaticMessageProvider(streamCallCount uint, inputs []*dynamic.Message) (*staticMessageProvider, error) {
	return &staticMessageProvider{
		streamCallCount: streamCallCount,
		inputs:          inputs,
		inputLen:        uint(len(inputs)),
	}, nil
}

func (m *staticMessageProvider) GetStreamMessage(parentCallData *CallData) (*dynamic.Message, error) {
	isLast := false
	if m.streamCallCount > 0 {
		if m.counter >= m.streamCallCount {
			return nil, ErrEndStream
		} else if m.indexCounter == m.inputLen {
			m.indexCounter = 0
		}

		isLast = m.counter == m.streamCallCount-1
	} else if m.counter >= m.inputLen {
		return nil, ErrEndStream
	} else if m.counter == m.inputLen-1 {
		isLast = true
	}

	payload := m.inputs[m.indexCounter]

	m.counter++
	m.indexCounter++

	var err error
	if err == nil && isLast {
		err = ErrLastMessage
	}

	return payload, err
}
