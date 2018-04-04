package grpcannon

import (
	"testing"

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

	t.Run("true when a slice of correct maps", func(t *testing.T) {
		m1 := make(map[string]interface{})
		m1["name"] = "bob"
		m1["foo"] = "far"

		m2 := make(map[string]interface{})
		m2["foo"] = 1
		m2["bar"] = 2

		s := []map[string]interface{}{m1, m2}
		res := isArrayData(s)
		assert.True(t, res)
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
