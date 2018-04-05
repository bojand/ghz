package helloworld

import (
	"fmt"
	"io"
	"sync"

	context "golang.org/x/net/context"
)

// Greeter implements the GreeterServer for tests
type Greeter struct {
	streamData []*HelloReply
	mutex      *sync.Mutex

	callCounts map[string]int
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	s.mutex.Lock()
	s.callCounts["unary"]++
	s.mutex.Unlock()

	return &HelloReply{Message: "Hello " + in.Name}, nil
}

// SayHellos lists all hellos
func (s *Greeter) SayHellos(req *HelloRequest, stream Greeter_SayHellosServer) error {
	s.mutex.Lock()
	s.callCounts["ss"]++
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
	s.callCounts["cs"]++
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
	s.callCounts["bidi"]++
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
	s.callCounts["unary"] = 0
	s.callCounts["ss"] = 0
	s.callCounts["cs"] = 0
	s.callCounts["bidi"] = 0
	s.mutex.Unlock()
}

// GetCount gets the count for specific call type
func (s *Greeter) GetCount(key string) int {
	s.mutex.Lock()
	val, ok := s.callCounts[key]
	s.mutex.Unlock()
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

	m := make(map[string]int)
	m["unary"] = 0
	m["ss"] = 0
	m["cs"] = 0
	m["bidi"] = 0

	return &Greeter{streamData: streamData, callCounts: m, mutex: &sync.Mutex{}}
}
