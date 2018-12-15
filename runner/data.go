package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// creates a message from a map
// marshal to JSON then use jsonb to marshal to message
// this way we follow protobuf more closely and allow cammelCase properties.
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

func createPayloads(data string, mtd *desc.MethodDescriptor) (*dynamic.Message, *[]*dynamic.Message, error) {
	md := mtd.GetInputType()
	var input *dynamic.Message
	var streamInput []*dynamic.Message

	if len(data) > 0 {
		if strings.IndexRune(data, '[') == 0 {
			dataArray := make([]map[string]interface{}, 5)
			err := json.Unmarshal([]byte(data), &dataArray)
			if err != nil {
				return nil, nil, fmt.Errorf("Error unmarshalling payload. Data: '%v' Error: %v", data, err.Error())
			}

			elems := len(dataArray)
			if elems > 0 {
				streamInput = make([]*dynamic.Message, elems)
			}
			for i, elem := range dataArray {
				elemMsg := dynamic.NewMessage(md)
				err := messageFromMap(elemMsg, &elem)
				if err != nil {
					return nil, nil, fmt.Errorf("Error creating message: %v", err.Error())
				}

				streamInput[i] = elemMsg
			}
		} else {
			input = dynamic.NewMessage(md)
			err := jsonpb.UnmarshalString(data, input)
			if err != nil {
				return nil, nil, fmt.Errorf("Error creating message from data. Data: '%v' Error: %v", data, err.Error())
			}
		}
	}

	if mtd.IsClientStreaming() && len(streamInput) == 0 && input != nil {
		streamInput = make([]*dynamic.Message, 1)
		streamInput[0] = input
		input = nil
	}

	if !mtd.IsClientStreaming() && input == nil && len(streamInput) > 0 {
		input = streamInput[0]
		streamInput = nil
	}

	return input, &streamInput, nil
}

func createPayloadsFromBin(binData []byte, mtd *desc.MethodDescriptor) (*dynamic.Message, *[]*dynamic.Message, error) {
	md := mtd.GetInputType()
	input := dynamic.NewMessage(md)
	streamInput := make([]*dynamic.Message, 1)

	err := proto.Unmarshal(binData, input)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating message from binary data: %v", err.Error())
	}

	if mtd.IsClientStreaming() && input != nil {
		streamInput = make([]*dynamic.Message, 1)
		streamInput[0] = input
		input = nil
	}

	return input, &streamInput, nil
}
