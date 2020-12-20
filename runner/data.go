package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc/metadata"
)

// TODO fix this so it can be a public API via interface
// TODO add tests
type dataProvider struct {
	binary   bool
	data     []byte
	metadata []byte
	mtd      *desc.MethodDescriptor
	dataFunc BinaryDataFunc

	arrayJSONData []string
	hasActions    bool

	// cached messages only for binary
	mutex          sync.RWMutex
	cachedMessages []*dynamic.Message
}

func newDataProvider(mtd *desc.MethodDescriptor,
	binary bool, dataFunc BinaryDataFunc,
	data, metadata []byte) (*dataProvider, error) {

	dp := dataProvider{
		binary:         binary,
		dataFunc:       dataFunc,
		mtd:            mtd,
		data:           data,
		metadata:       metadata,
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
	ctd := newCallData(mtd, nil, "", 0)
	ha, err := ctd.hasAction(string(dp.data))
	if err != nil {
		return nil, err
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

	if !dp.mtd.IsClientStreaming() {
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
			cdata, err := ctd.executeData(string(inputData))
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

func (dp *dataProvider) getMetadataForCall(ctd *CallData) (*metadata.MD, error) {
	mdMap, err := ctd.executeMetadata(string(dp.metadata))
	if err != nil {
		return nil, err
	}

	var reqMD *metadata.MD
	if len(mdMap) > 0 {
		md := metadata.New(mdMap)
		reqMD = &md
	} else {
		reqMD = &metadata.MD{}
	}

	return reqMD, nil
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
