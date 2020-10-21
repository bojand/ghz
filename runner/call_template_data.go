package runner

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/jhump/protoreflect/desc"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// call template data
type callTemplateData struct {
	WorkerID           string // unique worker ID
	RequestNumber      int64  // unique incremented request number for each request
	FullyQualifiedName string // fully-qualified name of the method call
	MethodName         string // shorter call method name
	ServiceName        string // the service name
	InputName          string // name of the input message type
	OutputName         string // name of the output message type
	IsClientStreaming  bool   // whether this call is client streaming
	IsServerStreaming  bool   // whether this call is server streaming
	Timestamp          string // timestamp of the call in RFC3339 format
	TimestampUnix      int64  // timestamp of the call as unix time in seconds
	TimestampUnixMilli int64  // timestamp of the call as unix time in milliseconds
	TimestampUnixNano  int64  // timestamp of the call as unix time in nanoseconds
	UUID               string // generated UUIDv4 for each call

	templateFuncs template.FuncMap
}

var tmplFuncMap = template.FuncMap{
	"newUUID":      newUUID,
	"randomString": randomString,
}

// newCallTemplateData returns new call template data
func newCallTemplateData(
	mtd *desc.MethodDescriptor,
	funcs template.FuncMap,
	workerID string, reqNum int64) *callTemplateData {
	now := time.Now()
	newUUID, _ := uuid.NewRandom()

	fns := make(template.FuncMap, len(funcs)+2)
	for k, v := range tmplFuncMap {
		fns[k] = v
	}

	if len(funcs) > 0 {
		for k, v := range funcs {
			fns[k] = v
		}
	}

	return &callTemplateData{
		WorkerID:           workerID,
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
		TimestampUnixMilli: now.UnixNano() / 1000000,
		TimestampUnixNano:  now.UnixNano(),
		UUID:               newUUID.String(),
		templateFuncs:      fns,
	}
}

func (td *callTemplateData) execute(data string) (*bytes.Buffer, error) {
	t := template.Must(template.New("call_template_data").Funcs(td.templateFuncs).Parse(data))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, td)
	return &tpl, err
}

func (td *callTemplateData) executeData(data string) ([]byte, error) {
	if len(data) > 0 {
		input := []byte(data)
		tpl, err := td.execute(data)
		if err == nil {
			input = tpl.Bytes()
		}

		return input, nil
	}

	return []byte{}, nil
}

func (td *callTemplateData) executeMetadata(metadata string) (map[string]string, error) {
	var md map[string]string

	if len(metadata) > 0 {
		input := []byte(metadata)
		tpl, err := td.execute(metadata)
		if err == nil {
			input = tpl.Bytes()
		}

		err = json.Unmarshal(input, &md)
		if err != nil {
			return nil, err
		}
	}

	return md, nil
}

// Same as executeMetadata, but this method ensures that the input metadata JSON string is always
// an array. If the input is an object, but not an array, it's converted to an array.
func (td *callTemplateData) executeMetadataArray(metadata string) ([]map[string]string, error) {
	var mdArray []map[string]string
	var metadataSanitized = strings.TrimSpace(metadata)

	// If the input is an object and not an array, we ensure we always work with an array
	if !strings.HasPrefix(metadataSanitized, "[") && !strings.HasSuffix(metadataSanitized, "]") {
		metadata = "[" + metadataSanitized + "]"
	}

	if len(metadata) > 0 {
		input := []byte(metadata)
		tpl, err := td.execute(metadata)
		if err == nil {
			input = tpl.Bytes()
		}

		err = json.Unmarshal(input, &mdArray)
		if err != nil {
			return nil, err
		}
	}

	return mdArray, nil
}

func newUUID() string {
	newUUID, _ := uuid.NewRandom()
	return newUUID.String()
}

const maxLen = 16
const minLen = 2

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomString(length int) string {
	if length <= 0 {
		length = seededRand.Intn(maxLen-minLen+1) + minLen
	}

	return stringWithCharset(length, charset)
}
