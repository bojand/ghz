package grpcannon

import (
	"errors"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// checks if data is an array / slice of maps
func isArrayData(data interface{}) bool {
	arrData, isArrData := data.([]interface{})
	if !isArrData {
		return false
	}

	for _, elem := range arrData {
		if !isMapData(elem) {
			return false
		}
	}

	return true
}

// checks that a data object is a map with string for keys
func isMapData(data interface{}) bool {
	_, isArrData := data.(map[string]interface{})
	return isArrData
}

// creates a message from a map
func messageFromMap(input *dynamic.Message, data *map[string]interface{}) error {
	for k, v := range *data {
		err := input.TrySetFieldByName(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPayloads(data interface{}, mtd *desc.MethodDescriptor) (*dynamic.Message, *[]*dynamic.Message, error) {
	md := mtd.GetInputType()
	var input *dynamic.Message
	var streamInput []*dynamic.Message

	// payload
	if isArrayData(data) {
		data := data.([]interface{})
		elems := len(data)
		if elems > 0 {
			streamInput = make([]*dynamic.Message, elems)
		}
		for i, elem := range data {
			o := elem.(map[string]interface{})
			elemMsg := dynamic.NewMessage(md)
			err := messageFromMap(elemMsg, &o)
			if err != nil {
				return nil, nil, err
			}

			streamInput[i] = elemMsg
		}
	} else if isMapData(data) {
		input = dynamic.NewMessage(md)
		data := data.(map[string]interface{})
		err := messageFromMap(input, &data)
		if err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, errors.New("Unsupported type for Data")
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
