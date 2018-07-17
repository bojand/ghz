package ghz

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/jhump/protoreflect/desc"
)

type callTemplateData struct {
	RequestNumber      int64
	FullyQualifiedName string
	MethodName         string
	ServiceName        string
	InputName          string
	OutputName         string
	IsClientStreaming  bool
	IsServerStreaming  bool
}

// newCallTemplateData returns new call template data
func newCallTemplateData(mtd *desc.MethodDescriptor, reqNum int64) *callTemplateData {
	return &callTemplateData{
		RequestNumber:      reqNum,
		FullyQualifiedName: mtd.GetFullyQualifiedName(),
		MethodName:         mtd.GetName(),
		ServiceName:        mtd.GetService().GetName(),
		InputName:          mtd.GetInputType().GetName(),
		OutputName:         mtd.GetOutputType().GetName(),
		IsClientStreaming:  mtd.IsClientStreaming(),
		IsServerStreaming:  mtd.IsServerStreaming(),
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
