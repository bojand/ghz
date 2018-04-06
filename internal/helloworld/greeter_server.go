package helloworld

import (
	"fmt"
	"io"
	"sync"

	context "golang.org/x/net/context"
)

// CallType represents one of the gRPC call types:
// unary, client streaming, server streaming, bidi
type CallType string

// Unary is a uniry call
var Unary CallType = "unary"

// ClientStream is a client streaming call
var ClientStream CallType = "cs"

// ServerStream is a server streaming call
var ServerStream CallType = "ss"

// Bidi is a bidi / duplex call
var Bidi CallType = "bidi"

// Greeter implements the GreeterServer for tests
type Greeter struct {
	streamData []*HelloReply
	mutex      *sync.Mutex

	callCounts map[CallType]int
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	s.mutex.Lock()
	s.callCounts[Unary]++
	s.mutex.Unlock()

	return &HelloReply{Message: "Hello " + in.Name}, nil
}

// SayHellos lists all hellos
func (s *Greeter) SayHellos(req *HelloRequest, stream Greeter_SayHellosServer) error {
	s.mutex.Lock()
	s.callCounts[ServerStream]++
	s.mutex.Unlock()

	for _, msg := range s.streamData {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

// SayHelloCS is client streaming handler
func (s *Greeter) SayHelloCS(stream Greeter_SayHelloCSServer) error {
	s.mutex.Lock()
	s.callCounts[ClientStream]++
	s.mutex.Unlock()

	msgCount := 0

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			msgStr := fmt.Sprintf("Hello count: %d", msgCount)
			return stream.SendAndClose(&HelloReply{Message: msgStr})
		}
		if err != nil {
			return err
		}
		msgCount++
	}
}

// SayHelloBidi duplex call handler
func (s *Greeter) SayHelloBidi(stream Greeter_SayHelloBidiServer) error {
	s.mutex.Lock()
	s.callCounts[Bidi]++
	s.mutex.Unlock()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		msg := "Hello " + in.Name
		if err := stream.Send(&HelloReply{Message: msg}); err != nil {
			return err
		}
	}
}

// ResetCounters resets the call counts
func (s *Greeter) ResetCounters() {
	s.mutex.Lock()
	s.callCounts[Unary] = 0
	s.callCounts[ServerStream] = 0
	s.callCounts[ClientStream] = 0
	s.callCounts[Bidi] = 0
	s.mutex.Unlock()
}

// GetCount gets the count for specific call type
func (s *Greeter) GetCount(key CallType) int {
	val, ok := s.callCounts[key]
	if ok {
		return val
	}
	return -1
}

// NewGreeter creates new greeter server
func NewGreeter() *Greeter {
	streamData := []*HelloReply{
		&HelloReply{Message: "Hello Bob"},
		&HelloReply{Message: "Hello Kate"},
		&HelloReply{Message: "Hello Jim"},
		&HelloReply{Message: "Hello Sara"},
	}

	m := make(map[CallType]int)
	m[Unary] = 0
	m[ServerStream] = 0
	m[ClientStream] = 0
	m[Bidi] = 0

	return &Greeter{streamData: streamData, callCounts: m, mutex: &sync.Mutex{}}
}
