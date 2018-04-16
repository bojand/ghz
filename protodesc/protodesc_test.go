package protodesc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtodesc_GetMethodDescFromProto(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		md, err := GetMethodDescFromProto("pkg.Call", "invalid.proto", []string{})
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid call symbol", func(t *testing.T) {
		md, err := GetMethodDescFromProto("pkg.Call", "../testdata/greeter.proto", []string{})
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid package", func(t *testing.T) {
		md, err := GetMethodDescFromProto("helloworld.pkg.SayHello", "../testdata/greeter.proto", []string{})
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid method", func(t *testing.T) {
		md, err := GetMethodDescFromProto("helloworld.Greeter.Foo", "../testdata/greeter.proto", []string{})
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("valid symbol", func(t *testing.T) {
		md, err := GetMethodDescFromProto("helloworld.Greeter.SayHello", "../testdata/greeter.proto", []string{})
		assert.NoError(t, err)
		assert.NotNil(t, md)
	})

	t.Run("valid symbol slashes", func(t *testing.T) {
		md, err := GetMethodDescFromProto("helloworld.Greeter/SayHello", "../testdata/greeter.proto", []string{})
		assert.NoError(t, err)
		assert.NotNil(t, md)
	})
}

func TestProtodesc_GetMethodDescFromProtoSet(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("pkg.Call", "invalid.protoset")
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid call symbol", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("pkg.Call", "../testdata/bundle.protoset")
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid package", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("helloworld.pkg.SayHello", "../testdata/bundle.protoset")
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("invalid method", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("helloworld.Greeter.Foo", "../testdata/bundle.protoset")
		assert.Error(t, err)
		assert.Nil(t, md)
	})

	t.Run("valid symbol", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("helloworld.Greeter.SayHello", "../testdata/bundle.protoset")
		assert.NoError(t, err)
		assert.NotNil(t, md)
	})

	t.Run("valid symbol proto 2", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("cap.Capper.Cap", "../testdata/bundle.protoset")
		assert.NoError(t, err)
		assert.NotNil(t, md)
	})

	t.Run("valid symbol slashes", func(t *testing.T) {
		md, err := GetMethodDescFromProtoSet("helloworld.Greeter/SayHello", "../testdata/bundle.protoset")
		assert.NoError(t, err)
		assert.NotNil(t, md)
	})
}
