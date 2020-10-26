package runner

import (
	"strings"
	"testing"
	"text/template"

	"github.com/bojand/ghz/protodesc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCallData_New(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "../testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallData(md, nil, "worker_id_123", 100)

	assert.NotNil(t, ctd)
	assert.Equal(t, "worker_id_123", ctd.WorkerID)
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
	assert.NotZero(t, ctd.TimestampUnixMilli)
	assert.NotZero(t, ctd.TimestampUnixNano)
	assert.Equal(t, ctd.TimestampUnix, ctd.TimestampUnixMilli/1000)
	assert.NotEmpty(t, ctd.UUID)
	assert.Equal(t, 36, len(ctd.UUID))
}

func TestCallData_ExecuteData(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "../testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallData(md, nil, "worker_id_123", 200)

	assert.NotNil(t, ctd)

	var tests = []struct {
		name        string
		in          string
		expected    []byte
		expectError bool
	}{
		{"no template",
			`{"name":"bob"}`,
			[]byte(`{"name":"bob"}`),
			false,
		},
		{"with template",
			`{"name":"{{.WorkerID}} {{.RequestNumber}} bob {{.FullyQualifiedName}} {{.MethodName}} {{.ServiceName}} {{.InputName}} {{.OutputName}} {{.IsClientStreaming}} {{.IsServerStreaming}}"}`,
			[]byte(`{"name":"worker_id_123 200 bob helloworld.Greeter.SayHello SayHello Greeter HelloRequest HelloReply false false"}`),
			false,
		},
		{"with unknown action",
			`{"name":"asdf {{.Something}} {{.MethodName}} bob"}`,
			[]byte(`{"name":"asdf {{.Something}} {{.MethodName}} bob"}`),
			false,
		},
		{"with empty string",
			"",
			[]byte{},
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

func TestCallData_ExecuteMetadata(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "../testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallData(md, nil, "worker_id_123", 200)

	assert.NotNil(t, ctd)

	var tests = []struct {
		name        string
		in          string
		expected    interface{}
		expectError bool
	}{
		{"no template",
			`{"trace_id":"asdf"}`,
			map[string]string{"trace_id": "asdf"},
			false,
		},
		{"with template",
			`{"trace_id":"{{.RequestNumber}} asdf {{.FullyQualifiedName}} {{.MethodName}} {{.ServiceName}} {{.InputName}} {{.OutputName}} {{.IsClientStreaming}} {{.IsServerStreaming}}"}`,
			map[string]string{"trace_id": "200 asdf helloworld.Greeter.SayHello SayHello Greeter HelloRequest HelloReply false false"},
			false,
		},
		{"with unknown action",
			`{"trace_id":"asdf {{.Something}} {{.MethodName}} bob"}`,
			map[string]string{"trace_id": "asdf {{.Something}} {{.MethodName}} bob"},
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

func TestCallTemplateData_ExecuteFuncs(t *testing.T) {
	md, err := protodesc.GetMethodDescFromProto("helloworld.Greeter/SayHello", "../testdata/greeter.proto", []string{})
	assert.NoError(t, err)
	assert.NotNil(t, md)

	ctd := newCallData(md, nil, "worker_id_123", 200)

	assert.NotNil(t, ctd)

	t.Run("newUUID", func(t *testing.T) {

		// no template
		r, err := ctd.executeData(`{"trace_id":"asdf"}`)
		assert.NoError(t, err)
		assert.Equal(t, `{"trace_id":"asdf"}`, string(r))

		rm, err := ctd.executeMetadata(`{"trace_id":"asdf"}`)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"trace_id": "asdf"}, rm)

		// new uuid
		r, err = ctd.executeData(`{"trace_id":"{{newUUID}}"}`)
		assert.NoError(t, err)
		rs := strings.Replace(string(r), `{"trace_id":"`, "", -1)
		rs = strings.Replace(rs, `"}`, "", -1)
		assert.NotEmpty(t, rs)
		parsed, err := uuid.Parse(rs)
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)

		rm2, err := ctd.executeMetadata(`{"trace_id":"{{newUUID}}"}`)
		assert.NoError(t, err)
		rs2 := rm2["trace_id"]
		assert.NotEmpty(t, rs)
		assert.NotEqual(t, rs, rs2)
		parsed, err = uuid.Parse(rs2)
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)

		rm3, err := ctd.executeMetadata(`{"span_id":"{{newUUID}}","trace_id":"{{newUUID}}"}`)
		assert.NoError(t, err)
		assert.NotEmpty(t, rs)
		assert.NotEqual(t, rm3["span_id"], rs2)
		assert.NotEqual(t, rm3["span_id"], rs)
		assert.NotEqual(t, rm3["trace_id"], rs2)
		assert.NotEqual(t, rm3["trace_id"], rs)
		assert.NotEqual(t, rm3["trace_id"], rm3["span_id"])
		parsed, err = uuid.Parse(rm3["span_id"])
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)
		parsed, err = uuid.Parse(rm3["trace_id"])
		assert.NoError(t, err)
		assert.NotEmpty(t, parsed)
	})

	t.Run("randomString", func(t *testing.T) {

		// no template
		r, err := ctd.executeData(`{"trace_id":"asdf"}`)
		assert.NoError(t, err)
		assert.Equal(t, `{"trace_id":"asdf"}`, string(r))

		rm, err := ctd.executeMetadata(`{"trace_id":"asdf"}`)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"trace_id": "asdf"}, rm)

		// default length when 0
		r, err = ctd.executeData(`{"trace_id":"{{randomString 0}}"}`)
		assert.NoError(t, err)
		rs := strings.Replace(string(r), `{"trace_id":"`, "", -1)
		rs = strings.Replace(rs, `"}`, "", -1)
		assert.NotEmpty(t, rs)
		assert.True(t, len(rs) >= 2)
		assert.True(t, len(rs) <= 16)

		// default length when -1
		r, err = ctd.executeData(`{"trace_id":"{{randomString -1}}"}`)
		assert.NoError(t, err)
		rs2 := strings.Replace(string(r), `{"trace_id":"`, "", -1)
		rs2 = strings.Replace(rs2, `"}`, "", -1)
		assert.NotEmpty(t, rs2)
		assert.NotEqual(t, rs, rs2)
		assert.True(t, len(rs2) >= 2)
		assert.True(t, len(rs2) <= 16)

		// specific length
		r, err = ctd.executeData(`{"trace_id":"{{randomString 10}}"}`)
		assert.NoError(t, err)
		rs = strings.Replace(string(r), `{"trace_id":"`, "", -1)
		rs = strings.Replace(rs, `"}`, "", -1)
		assert.NotEmpty(t, rs)
		assert.Len(t, rs, 10)

		rm, err = ctd.executeMetadata(`{"trace_id":"{{randomString 0}}"}`)
		assert.NoError(t, err)
		assert.True(t, len(rm["trace_id"]) >= 2)
		assert.True(t, len(rm["trace_id"]) <= 16)

		rm, err = ctd.executeMetadata(`{"span_id":"{{randomString -1}}","trace_id":"{{randomString 0}}"}`)
		assert.NoError(t, err)
		assert.True(t, len(rm["trace_id"]) >= 2)
		assert.True(t, len(rm["trace_id"]) <= 16)
		assert.True(t, len(rm["span_id"]) >= 2)
		assert.True(t, len(rm["span_id"]) <= 16)
		assert.NotEqual(t, rm["trace_id"], rm["span_id"])

		rm, err = ctd.executeMetadata(`{"span_id":"{{randomString 12}}","trace_id":"{{randomString 12}}"}`)
		assert.NoError(t, err)
		assert.Len(t, rm["trace_id"], 12)
		assert.Len(t, rm["span_id"], 12)
		assert.NotEqual(t, rm["trace_id"], rs)
		assert.NotEqual(t, rm["trace_id"], rs2)
		assert.NotEqual(t, rm["trace_id"], rm["span_id"])
	})

	t.Run("custom functions", func(t *testing.T) {
		ctd = newCallData(md, template.FuncMap{
			"getSKU": func() string {
				return "custom-sku"
			},
			"newUUID": func() string {
				return "custom-uuid"
			},
		}, "worker_id_123", 200)

		r, err := ctd.executeData(`{"trace_id":"{{newUUID}}", "span_id":"{{getSKU}}"}`)
		assert.NoError(t, err)
		assert.Equal(t, `{"trace_id":"custom-uuid", "span_id":"custom-sku"}`, string(r))

		rm, err := ctd.executeMetadata(`{"span_id":"{{randomString 12}}","trace_id":"{{newUUID}}", "sku":"{{getSKU}}"}`)
		assert.NoError(t, err)
		assert.Len(t, rm["span_id"], 12)
		assert.Equal(t, "custom-uuid", rm["trace_id"])
		assert.Equal(t, "custom-sku", rm["sku"])
	})
}
