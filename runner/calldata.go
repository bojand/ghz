package runner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"strings"
	"text/template"
	"text/template/parse"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/uuid"
	"github.com/jhump/protoreflect/desc"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// CallData represents contextualized data available for templating
type CallData struct {
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

	t *template.Template
}

var tmplFuncMap = template.FuncMap{
	"newUUID":      newUUID,
	"randomString": randomString,
	"randomInt":    randomInt,
}

// newCallData returns new CallData
func newCallData(
	mtd *desc.MethodDescriptor,
	funcs template.FuncMap,
	workerID string, reqNum int64) *CallData {

	fns := make(template.FuncMap, len(funcs)+2)
	for k, v := range tmplFuncMap {
		fns[k] = v
	}

	for k, v := range sprig.FuncMap() {
		fns[k] = v
	}

	if len(funcs) > 0 {
		for k, v := range funcs {
			fns[k] = v
		}
	}

	t := template.New("call_template_data").Funcs(fns)

	now := time.Now()
	newUUID, _ := uuid.NewRandom()

	return &CallData{
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
		t:                  t,
	}
}

// Regenerate generates a new instance of call data from this parent instance
// The dynamic data like timestamps and UUIDs are re-filled
func (td *CallData) Regenerate() *CallData {
	now := time.Now()
	newUUID, _ := uuid.NewRandom()

	return &CallData{
		WorkerID:           td.WorkerID,
		RequestNumber:      td.RequestNumber,
		FullyQualifiedName: td.FullyQualifiedName,
		MethodName:         td.MethodName,
		ServiceName:        td.ServiceName,
		InputName:          td.InputName,
		OutputName:         td.OutputName,
		IsClientStreaming:  td.IsClientStreaming,
		IsServerStreaming:  td.IsServerStreaming,
		Timestamp:          now.Format(time.RFC3339),
		TimestampUnix:      now.Unix(),
		TimestampUnixMilli: now.UnixNano() / 1000000,
		TimestampUnixNano:  now.UnixNano(),
		UUID:               newUUID.String(),
		t:                  td.t,
	}
}

func (td *CallData) execute(data string) (*bytes.Buffer, error) {
	t, err := td.t.Parse(data)
	if err != nil {
		return nil, err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, td)
	return &tpl, err
}

// This is hacky.
// See https://golang.org/pkg/text/template/#Template
// The *parse.Tree field is exported only for use by html/template
// and should be treated as unexported by all other clients.
func (td *CallData) hasAction(data string) (bool, error) {
	t, err := td.t.Parse(data)
	if err != nil {
		return false, err
	}

	hasAction := hasAction(t.Tree.Root)
	return hasAction, nil
}

func hasAction(node parse.Node) bool {
	has := false
	if node.Type() == parse.NodeAction {
		return true
	} else if ln, ok := node.(*parse.ListNode); ok {
		for _, n := range ln.Nodes {
			v := hasAction(n)
			if !has && v {
				has = true
				break
			}
		}
	}

	return has
}

// ExecuteData applies the call data's parsed template and data string and returns the resulting buffer
func (td *CallData) ExecuteData(data string) ([]byte, error) {
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

func (td *CallData) executeMetadata(metadata string) (map[string]string, error) {
	var mdMap map[string]string

	if len(metadata) > 0 {
		input := []byte(metadata)
		tpl, err := td.execute(metadata)
		if err == nil {
			input = tpl.Bytes()
		}

		err = json.Unmarshal(input, &mdMap)
		if err != nil {
			return nil, err
		}

		for key, value := range mdMap {
			if strings.HasSuffix(key, "-bin") {
				decoded, err := base64.StdEncoding.DecodeString(value)
				if err != nil {
					return nil, err
				}
				mdMap[key] = string(decoded)
			}
		}
	}

	return mdMap, nil
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

func randomInt(min, max int) int {
	if min < 0 {
		min = 0
	}

	if max <= 0 {
		max = 1
	}

	return seededRand.Intn(max-min) + min
}
