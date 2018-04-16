package grpcannon

import (
	"testing"

	"github.com/bojand/grpcannon/protodesc"

	"github.com/stretchr/testify/assert"
)

func TestData_isArrayData(t *testing.T) {
	t.Run("false when nil", func(t *testing.T) {
		res := isArrayData(nil)
		assert.False(t, res)
	})

	t.Run("false when a string", func(t *testing.T) {
		res := isArrayData("asdf")
		assert.False(t, res)
	})

	t.Run("false when a number", func(t *testing.T) {
		res := isArrayData(5)
		assert.False(t, res)
	})

	t.Run("false when an empty map", func(t *testing.T) {
		m := make(map[string]interface{})
		res := isArrayData(m)
		assert.False(t, res)
	})

	t.Run("false when a map", func(t *testing.T) {
		m := make(map[string]interface{})
		m["name"] = "bob"
		res := isArrayData(m)
		assert.False(t, res)
	})

	t.Run("false when an empty slice", func(t *testing.T) {
		res := isArrayData([]string{})
		assert.False(t, res)
	})

	t.Run("false when a slice of strings", func(t *testing.T) {
		res := isArrayData([]string{"foo", "bar"})
		assert.False(t, res)
	})

	t.Run("false when a slice of maps as we require slice of ifaces which are maps", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		s := []map[string]interface{}{m1, m2}
		res := isArrayData(s)
		assert.False(t, res)
	})

	t.Run("true when a slice of correct maps as interfaces", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		s := []interface{}{m1, m2}
		res := isArrayData(s)
		assert.True(t, res)
	})

	t.Run("false when a slice of containing incorrect element", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		s := []interface{}{m1, m2, "some string"}
		res := isArrayData(s)
		assert.False(t, res)
	})

	t.Run("false when a slice of containing incorrect map element", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		m3 := make(map[int]interface{})
		m3[5] = "foo"
		m3[7] = "bar"

		s := []interface{}{m1, m2, m3}
		res := isArrayData(s)
		assert.False(t, res)
	})
}

func TestData_isMapData(t *testing.T) {
	t.Run("false when nil", func(t *testing.T) {
		res := isMapData(nil)
		assert.False(t, res)
	})

	t.Run("false when a string", func(t *testing.T) {
		res := isMapData("asdf")
		assert.False(t, res)
	})

	t.Run("false when a number", func(t *testing.T) {
		res := isMapData(5)
		assert.False(t, res)
	})

	t.Run("false when a map of invalid keys", func(t *testing.T) {
		m := make(map[int]interface{})
		res := isMapData(m)
		assert.False(t, res)
	})

	t.Run("true when a valid map", func(t *testing.T) {
		m := make(map[string]interface{})
		m["name"] = "bob"
		res := isMapData(m)
		assert.True(t, res)
	})

	t.Run("true when a valid empty map", func(t *testing.T) {
		m := make(map[string]interface{})
		res := isMapData(m)
		assert.True(t, res)
	})

	t.Run("false when a slice of valid maps", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		s := []map[string]interface{}{m1, m2}
		res := isMapData(s)
		assert.False(t, res)
	})
}

func TestData_createPayloads(t *testing.T) {
	mtdUnary, err := protodesc.GetMethodDesc(
		"helloworld.Greeter.SayHello",
		"./testdata/greeter.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdUnary)

	mtdClientStreaming, err := protodesc.GetMethodDesc(
		"helloworld.Greeter.SayHelloCS",
		"./testdata/greeter.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdClientStreaming)

	mtdTestUnary, err := protodesc.GetMethodDesc(
		"data.DataTestService.TestCall",
		"./testdata/data.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdTestUnary)

	mtdTestUnaryTwo, err := protodesc.GetMethodDesc(
		"data.DataTestService.TestCallTwo",
		"./testdata/data.proto",
		nil)

	assert.NoError(t, err)
	assert.NotNil(t, mtdTestUnaryTwo)

	t.Run("fail when nil", func(t *testing.T) {
		single, streaming, err := createPayloads(nil, mtdUnary)
		assert.Error(t, err)
		assert.Nil(t, single)
		assert.Nil(t, streaming)
	})

	t.Run("fail for invalid data shape", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["unknown"] = "field"

		single, streaming, err := createPayloads(m1, mtdUnary)
		assert.Error(t, err)
		assert.Nil(t, single)
		assert.Nil(t, streaming)
	})

	t.Run("create single object from map for unary", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		single, streaming, err := createPayloads(m1, mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create array from map for client streaming", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"

		single, streaming, err := createPayloads(m1, mtdClientStreaming)
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

		single, streaming, err := createPayloads(s, mtdClientStreaming)
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

		single, streaming, err := createPayloads(s, mtdClientStreaming)
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

		single, streaming, err := createPayloads(s, mtdUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with camelCase property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["paramOne"] = "bob"

		single, streaming, err := createPayloads(m1, mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with snake_case property", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["param_one"] = "bob"

		single, streaming, err := createPayloads(m1, mtdTestUnary)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with nested camelCase property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["paramOne"] = "bob"

		m1 := make(map[string]interface{})
		m1["nestedProp"] = inner

		single, streaming, err := createPayloads(m1, mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})

	t.Run("create single object from map for unary with nested snake_case property", func(t *testing.T) {
		inner := make(map[string]interface{})
		inner["param_one"] = "bob"

		m1 := make(map[string]interface{})
		m1["nested_prop"] = inner

		single, streaming, err := createPayloads(m1, mtdTestUnaryTwo)
		assert.NoError(t, err)
		assert.NotNil(t, single)
		assert.Empty(t, streaming)
	})
}
