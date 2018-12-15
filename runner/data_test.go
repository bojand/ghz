package runner

import (
	"encoding/json"
	"testing"

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

	t.Run("get nil, emtpy when empty", func(t *testing.T) {
		single, streaming, err := createPayloads("", mtdUnary)
		assert.NoError(t, err)
		assert.Nil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("fail for invalid data shape", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["unknown"] = "field"

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdUnary)
		assert.Error(t, err)
		assert.Nil(t, single)
		assert.Nil(t, streaming)
	})

	t.Run("create single object from map for unary", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create array from map for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdClientStreaming)
		assert.NoError(t, err)
		assert.Nil(t, single)
		assert.NotNil(t, streaming)
		assert.Len(t, *streaming, 1)
	})

	t.Run("create slice of messages from slice for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "kate"

		s := []interface{}{m1, m2}

		jsonData, _ := json.Marshal(s)

		single, streaming, err := createPayloads(string(jsonData), mtdClientStreaming)
		assert.NoError(t, err)
		assert.Nil(t, single)
		assert.NotNil(t, streaming)
		assert.Len(t, *streaming, 2)
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

		single, streaming, err := createPayloads(string(jsonData), mtdClientStreaming)
		assert.Error(t, err)
		assert.Nil(t, single)
		assert.Nil(t, streaming)
	})

	t.Run("get object for slice and unary", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		m2 := make(map[string]interface{})
		m2["name"] = "kate"

		m3 := make(map[string]interface{})
		m3["name"] = "Jim"

		s := []interface{}{m1, m2, m3}

		jsonData, _ := json.Marshal(s)

		single, streaming, err := createPayloads(string(jsonData), mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with camelCase property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["paramOne"] = "bob"

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with snake_case property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["param_one"] = "bob"

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with nested camelCase property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["paramOne"] = "bob"

		m1 := make(map[string]interface{})
		m1["nestedProp"] = inner

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with nested snake_case property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["param_one"] = "bob"

		m1 := make(map[string]interface{})
		m1["nested_prop"] = inner

		jsonData, _ := json.Marshal(m1)

		single, streaming, err := createPayloads(string(jsonData), mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})
}
