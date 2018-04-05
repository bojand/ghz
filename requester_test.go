package grpcannon

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bojand/grpcannon/internal/helloworld"
	"github.com/bojand/grpcannon/protodesc"
	"google.golang.org/grpc"
)

const port = ":50051"
const localhost = "0.0.0.0:50051"

func startServer() (*helloworld.Greeter, *grpc.Server, error) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, err
	}

	s := grpc.NewServer()

	gs := helloworld.NewGreeter()
	helloworld.RegisterGreeterServer(s, gs)
	// reflection.Register(s)
	go func() {
		s.Serve(lis)
	}()
	return gs, s, err
}

func TestRequestUnary(t *testing.T) {
	fmt.Println("starting server")
	gs, s, err := startServer()
	fmt.Println("server started")
	assert.NoError(t, err)
	defer s.Stop()

	md, err := protodesc.GetMethodDesc("helloworld.Greeter.SayHello", "./testdata/greeter.proto", []string{})

	data := make(map[string]interface{})
	data["name"] = "bob"

	t.Run("test N", func(t *testing.T) {
		fmt.Printf("testing N")
		reqr, err := New(md, &Options{
			Host:        localhost,
			N:           20,
			C:           2,
			Timeout:     20,
			DialTimtout: 20,
			Data:        data,
		})
		assert.NoError(t, err)

		report, err := reqr.Run()
		assert.NoError(t, err)
		assert.NotNil(t, report)
		count := gs.CallCounts["unary"]
		assert.Equal(t, 20, count)
	})
}
