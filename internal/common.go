package internal

import (
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/bojand/ghz/internal/gtime"
	"github.com/bojand/ghz/internal/helloworld"
	"github.com/bojand/ghz/internal/sleep"
	"github.com/bojand/ghz/internal/wrapped"
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

// StartSleepServer starts the sleep test server
func StartSleepServer(secure bool) (*sleep.SleepService, *grpc.Server, error) {
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

	ss := sleep.SleepService{}
	sleep.RegisterSleepServiceServer(s, &ss)
	reflection.Register(s)

	TestPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	TestLocalhost = "localhost:" + TestPort

	go func() {
		_ = s.Serve(lis)
	}()

	return &ss, s, err
}

// StartWrappedServer starts the wrapped test server
func StartWrappedServer(secure bool) (*wrapped.WrappedService, *grpc.Server, error) {
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

	ws := wrapped.WrappedService{}
	wrapped.RegisterWrappedServiceServer(s, &ws)
	reflection.Register(s)

	TestPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	TestLocalhost = "localhost:" + TestPort

	go func() {
		_ = s.Serve(lis)
	}()

	return &ws, s, err
}

// StartTimeServer starts the wrapped test server
func StartTimeServer(secure bool) (*gtime.TimeService, *grpc.Server, error) {
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

	gs := gtime.TimeService{}
	gtime.RegisterTimeServiceServer(s, &gs)
	reflection.Register(s)

	TestPort = strconv.Itoa(lis.Addr().(*net.TCPAddr).Port)
	TestLocalhost = "localhost:" + TestPort

	go func() {
		_ = s.Serve(lis)
	}()

	return &gs, s, err
}
