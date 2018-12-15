package runner

import (
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/bojand/ghz/internal/helloworld"
)

var port string
var localhost string

func startServer(secure bool) (*helloworld.Greeter, *grpc.Server, error) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, err
	}

	var opts []grpc.ServerOption

	if secure {
		creds, err := credentials.NewServerTLSFromFile("../testdata/localhost.crt", "../testdata/localhost.key")
		if err != nil {
			return nil, nil, err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	s := grpc.NewServer(opts...)

	gs := helloworld.NewGreeter()
	helloworld.RegisterGreeterServer(s, gs)
	// reflection.Register(s)

	port = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	localhost = "localhost:" + port

	go func() {
		s.Serve(lis)
	}()

	return gs, s, err
}
