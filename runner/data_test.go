package runner

import (
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/bojand/ghz/internal/helloworld"
	"github.com/bojand/ghz/protodesc"

	"github.com/stretchr/testify/assert"
)

func TestData_createPayloads(t *testing.T) {
	mtdUnary, err := protodesc.GetMethodDescFromProto(
		"helloworld.Greeter.SayHello",
		"../testdata/greeter.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdUnary)

	mtdClientStreaming, err := protodesc.GetMethodDescFromProto(
		"helloworld.Greeter.SayHelloCS",
		"../testdata/greeter.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdClientStreaming)

	mtdTestUnary, err := protodesc.GetMethodDescFromProto(
		"data.DataTestService.TestCall",
		"../testdata/data.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdTestUnary)

	mtdTestUnaryTwo, err := protodesc.GetMethodDescFromProto(
		"data.DataTestService.TestCallTwo",
		"../testdata/data.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdTestUnaryTwo)

	mtdWrapped, err := protodesc.GetMethodDescFromProto(
		"wrapped.WrappedService.GetMessage",
		"../testdata/wrapped.proto",
		[]string{"../testdata"})

	assert.NoError(t, err)
	assert.NotNil(t, mtdUnary)

	t.Run("get empty when empty", func(t *testing.T) {
		inputs, err := createPayloadsFromJSON("", mtdUnary)
		assert.NoError(t, err)
		assert.Empty(t, inputs)
	})

	t.Run("fail for invalid data shape", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["unknown"] = "field"

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdUnary)
		assert.Error(t, err)
		assert.Nil(t, inputs)
	})

	t.Run("create slice with single element from map for unary", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create slice with single element from map for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdClientStreaming)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create slice of messages from slice for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "kate"

		s := []interface{}{m1, m2}

		jsonData, _ := json.Marshal(s)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdClientStreaming)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 2)
	})

	t.Run("fail on invalid shape of data in slice for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "kate"

		m3 := make(map[string]interface{})
		m3["name"] = "Jim"
		m3["foo"] = "bar"

		s := []interface{}{m1, m2, m3}

		jsonData, _ := json.Marshal(s)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdClientStreaming)
		assert.Error(t, err)
		assert.Nil(t, inputs)
	})

	t.Run("create slice of messages from slice for unary", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "kate"

		m3 := make(map[string]interface{})
		m3["name"] = "Jim"

		s := []interface{}{m1, m2, m3}

		jsonData, _ := json.Marshal(s)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 3)
	})

	t.Run("create slice with single object from map for unary with camelCase property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["paramOne"] = "bob"

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create slice with single object from map for unary with snake_case property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["param_one"] = "bob"

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create slice with single object from map for unary with nested camelCase property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["paramOne"] = "bob"

		m1 := make(map[string]interface{})
		m1["nestedProp"] = inner

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create slice with single object from map for unary with nested snake_case property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["param_one"] = "bob"

		m1 := make(map[string]interface{})
		m1["nested_prop"] = inner

		jsonData, _ := json.Marshal(m1)

		inputs, err := createPayloadsFromJSON(string(jsonData), mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.NotNil(t, inputs[0])
	})

	t.Run("create data with well-known wrapped type", func(t *testing.T) {
		jsonData := `"foo"`

		inputs, err := createPayloadsFromJSON(jsonData, mtdWrapped)
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.EqualValues(t, "foo", inputs[0].GetFieldByName("value"))
	})

	t.Run("create slice from single message binary data", func(t *testing.T) {
		msg1 := &helloworld.HelloRequest{}
		msg1.Name = "bob"

		binData, err := proto.Marshal(msg1)
		assert.NoError(t, err)

		inputs, err := createPayloadsFromBin(binData, mtdUnary)

		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 1)
		assert.EqualValues(t, msg1.GetName(), inputs[0].GetFieldByName("name"))
	})

	t.Run("create slice from count-delimited binary data", func(t *testing.T) {
		msg1 := &helloworld.HelloRequest{}
		msg1.Name = "bob"
		msg2 := &helloworld.HelloRequest{}
		msg2.Name = "alice"

		buf := proto.Buffer{}
		_ = buf.EncodeMessage(msg1)
		_ = buf.EncodeMessage(msg2)

		inputs, err := createPayloadsFromBin(buf.Bytes(), mtdUnary)

		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 2)
		assert.EqualValues(t, msg1.GetName(), inputs[0].GetFieldByName("name"))
		assert.EqualValues(t, msg2.GetName(), inputs[1].GetFieldByName("name"))
	})

	t.Run("on empty binary data returns empty slice", func(t *testing.T) {
		buf := make([]byte, 0)
		inputs, err := createPayloadsFromBin(buf, mtdUnary)

		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Len(t, inputs, 0)
	})
}
