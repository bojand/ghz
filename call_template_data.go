package ghz

import (
	"bytes"
	"encoding/json"
	"text/template"
	"time"

	"github.com/jhump/protoreflect/desc"
)

// call template data
type callTemplateData struct {
	RequestNumber      int64  // unique incrememnted request number for each request
	FullyQualifiedName string // fully-qualified name of the method call
	MethodName         string // shorter call method name
	ServiceName        string // the service name
	InputName          string // name of the input message type
	OutputName         string // name of the output message type
	IsClientStreaming  bool   // whether this call is client streaming
	IsServerStreaming  bool   // whether this call is server streaming
	Timestamp          string // timestamp of the call in RFC3339 format
	TimestampUnix      int64  // timestamp of the call as unix time
}

// newCallTemplateData returns new call template data
func newCallTemplateData(mtd *desc.MethodDescriptor, reqNum int64) *callTemplateData {
	now := time.Now()

	return &callTemplateData{
		RequestNumber:      reqNum,
		FullyQualifiedName: mtd.GetFullyQualifiedName(),
		MethodName:         mtd.GetName(),
		ServiceName:        mtd.GetService().GetName(),
		InputName:          mtd.GetInputType().GetName(),
		OutputName:         mtd.GetOutputType().GetName(),
		IsClientStreaming:  mtd.IsClientStreaming(),
		IsServerStreaming:  mtd.IsServerStreaming(),
		Timestamp:          now.Format(time.RFC3339),
		TimestampUnix:      now.Unix(),
	}
}

func (td *callTemplateData) execute(data string) (*bytes.Buffer, error) {
	t := template.Must(template.New("call_template_data").Parse(data))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, td)
	return &tpl, err
}

func (td *callTemplateData) executeData(data string) (interface{}, error) {
	input := []byte(data)
	tpl, err := td.execute(data)
	if err == nil {
		input = tpl.Bytes()
	}

	var dataMap interface{}
	err = json.Unmarshal(input, &dataMap)
	if err != nil {
		return nil, err
	}

	return dataMap, nil
}

func (td *callTemplateData) executeMetadata(metadata string) (*map[string]string, error) {
	input := []byte(metadata)
	tpl, err := td.execute(metadata)
	if err == nil {
		input = tpl.Bytes()
	}

	var mdMap map[string]string
	err = json.Unmarshal(input, &mdMap)
	if err != nil {
		return nil, err
	}

	return &mdMap, nil
}
