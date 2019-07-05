package runner

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

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
