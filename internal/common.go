package internal

import (
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/bojand/ghz/internal/helloworld"
)

// TestPort is the port.
var TestPort string

// TestLocalhost is the localhost.
var TestLocalhost string

// StartServer starts the server.
//
// For testing only.
func StartServer(secure bool) (*helloworld.Greeter, *grpc.Server, error) {
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

	stats := helloworld.NewHWStats()

	opts = append(opts, grpc.StatsHandler(stats))

	s := grpc.NewServer(opts...)

	gs := helloworld.NewGreeter()
	helloworld.RegisterGreeterServer(s, gs)
	reflection.Register(s)

	gs.Stats = stats

	TestPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	TestLocalhost = "localhost:" + TestPort

	go func() {
		_ = s.Serve(lis)
	}()

	return gs, s, err
}
