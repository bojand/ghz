package ghz

import (
	"testing"

	"github.com/bojand/ghz/protodesc"
	"github.com/stretchr/testify/assert"
)

func TestCallTemplateData_New(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "./testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallTemplateData(md, 100)

	assert.NotNil(t, ctd)
	assert.Equal(t, int64(100), ctd.RequestNumber)
	assert.Equal(t, "helloworld.Greeter.SayHello", ctd.FullyQualifiedName)
	assert.Equal(t, "SayHello", ctd.MethodName)
	assert.Equal(t, "Greeter", ctd.ServiceName)
	assert.Equal(t, "HelloRequest", ctd.InputName)
	assert.Equal(t, "HelloReply", ctd.OutputName)
	assert.Equal(t, false, ctd.IsClientStreaming)
	assert.Equal(t, false, ctd.IsServerStreaming)
	assert.NotEmpty(t, ctd.Timestamp)
	assert.NotZero(t, ctd.TimestampUnix)
}

func TestCallTemplateData_ExecuteData(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "./testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallTemplateData(md, 200)

	assert.NotNil(t, ctd)

	var tests = []struct {
		name        string
		in          string
		expected    interface{}
		expectError bool
	}{
		{"no template",
			`{"name":"bob"}`,
			map[string]interface{}{"name": "bob"},
			false,
		},
		{"with template",
			`{"name":"{{.RequestNumber}} bob {{.FullyQualifiedName}} {{.MethodName}} {{.ServiceName}} {{.InputName}} {{.OutputName}} {{.IsClientStreaming}} {{.IsServerStreaming}}"}`,
			map[string]interface{}{"name": "200 bob helloworld.Greeter.SayHello SayHello Greeter HelloRequest HelloReply false false"},
			false,
		},
		{"with unknown action",
			`{"name":"asdf {{.Something}} {{.MethodName}} bob"}`,
			map[string]interface{}{"name": "asdf {{.Something}} {{.MethodName}} bob"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ctd.executeData(tt.in)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}

func TestCallTemplateData_ExecuteMetadata(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "./testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallTemplateData(md, 200)

	assert.NotNil(t, ctd)

	var tests = []struct {
		name        string
		in          string
		expected    interface{}
		expectError bool
	}{
		{"no template",
			`{"trace_id":"asdf"}`,
			&map[string]string{"trace_id": "asdf"},
			false,
		},
		{"with template",
			`{"trace_id":"{{.RequestNumber}} asdf {{.FullyQualifiedName}} {{.MethodName}} {{.ServiceName}} {{.InputName}} {{.OutputName}} {{.IsClientStreaming}} {{.IsServerStreaming}}"}`,
			&map[string]string{"trace_id": "200 asdf helloworld.Greeter.SayHello SayHello Greeter HelloRequest HelloReply false false"},
			false,
		},
		{"with unknown action",
			`{"trace_id":"asdf {{.Something}} {{.MethodName}} bob"}`,
			&map[string]string{"trace_id": "asdf {{.Something}} {{.MethodName}} bob"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ctd.executeMetadata(tt.in)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, r)
		})
	}
}
