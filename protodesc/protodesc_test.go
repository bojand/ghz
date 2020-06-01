package protodesc

import (
	"context"
	"testing"

	"github.com/bojand/ghz/internal"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
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

func TestParseServiceMethod(t *testing.T) {
	testParseServiceMethodSuccess(t, "package.Service.Method", "package.Service", "Method")
	testParseServiceMethodSuccess(t, ".package.Service.Method", "package.Service", "Method")
	testParseServiceMethodSuccess(t, "package.Service/Method", "package.Service", "Method")
	testParseServiceMethodSuccess(t, ".package.Service/Method", "package.Service", "Method")
	testParseServiceMethodSuccess(t, "Service.Method", "Service", "Method")
	testParseServiceMethodSuccess(t, ".Service.Method", "Service", "Method")
	testParseServiceMethodSuccess(t, "Service/Method", "Service", "Method")
	testParseServiceMethodSuccess(t, ".Service/Method", "Service", "Method")
	testParseServiceMethodError(t, "")
	testParseServiceMethodError(t, ".")
	testParseServiceMethodError(t, "package/Service/Method")
}

func testParseServiceMethodSuccess(t *testing.T, svcAndMethod string, expectedService string, expectedMethod string) {
	service, method, err := parseServiceMethod(svcAndMethod)
	assert.NoError(t, err)
	assert.Equal(t, expectedService, service)
	assert.Equal(t, expectedMethod, method)
}

func testParseServiceMethodError(t *testing.T, svcAndMethod string) {
	_, _, err := parseServiceMethod(svcAndMethod)
	assert.Error(t, err)
}

func TestProtodesc_GetMethodDescFromReflect(t *testing.T) {

	_, s, err := internal.StartServer(false)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer s.Stop()

	t.Run("test known call", func(t *testing.T) {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, internal.TestLocalhost, opts...)
		assert.NoError(t, err)

		md := make(metadata.MD)

		refCtx := metadata.NewOutgoingContext(ctx, md)

		refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(conn))

		mtd, err := GetMethodDescFromReflect("helloworld.Greeter.SayHello", refClient)
		assert.NoError(t, err)
		assert.NotNil(t, mtd)
		assert.Equal(t, "SayHello", mtd.GetName())
	})

	t.Run("test known call with /", func(t *testing.T) {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, internal.TestLocalhost, opts...)
		assert.NoError(t, err)

		md := make(metadata.MD)

		refCtx := metadata.NewOutgoingContext(ctx, md)

		refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(conn))

		mtd, err := GetMethodDescFromReflect("helloworld.Greeter/SayHello", refClient)
		assert.NoError(t, err)
		assert.NotNil(t, mtd)
		assert.Equal(t, "SayHello", mtd.GetName())
	})

	t.Run("test unknown known call", func(t *testing.T) {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, internal.TestLocalhost, opts...)
		assert.NoError(t, err)

		md := make(metadata.MD)

		refCtx := metadata.NewOutgoingContext(ctx, md)

		refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(conn))

		mtd, err := GetMethodDescFromReflect("helloworld.Greeter/SayHelloAsdf", refClient)
		assert.Error(t, err)
		assert.Nil(t, mtd)
	})
}
