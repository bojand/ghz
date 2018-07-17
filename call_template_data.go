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
	var err error

	t := template.New("call_template_data")
	t, err = t.Parse(data)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, td)
	return &tpl, err
}

func (td *callTemplateData) executeData(data string) (interface{}, error) {

	var tpl *bytes.Buffer
	var err error
	if tpl, err = td.execute(data); err != nil {
		return nil, err
	}

	var dataMap interface{}
	err = json.Unmarshal(tpl.Bytes(), &dataMap)
	if err != nil {
		return nil, err
	}

	return dataMap, nil
}

func (td *callTemplateData) executeMetadata(metadata string) (*map[string]string, error) {
	var tpl *bytes.Buffer
	var err error
	if tpl, err = td.execute(metadata); err != nil {
		return nil, err
	}

	var mdMap map[string]string
	err = json.Unmarshal(tpl.Bytes(), &mdMap)
	if err != nil {
		return nil, err
	}

	return &mdMap, nil
}
