package helloworld

import (
	"errors"
	"sync"

	context "golang.org/x/net/context"
)

// Greeter implements the GreeterServer for tests
type Greeter struct {
	streamData []*HelloReply
	mutex      *sync.Mutex

	CallCounts map[string]int
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	s.mutex.Lock()
	s.CallCounts["unary"]++
	s.mutex.Unlock()
	return &HelloReply{Message: "Hello " + in.Name}, nil
}

// SayHellos lists all hellos
func (s *Greeter) SayHellos(req *HelloRequest, stream Greeter_SayHellosServer) error {
	s.mutex.Lock()
	s.CallCounts["ss"]++
	s.mutex.Unlock()
	for _, message := range s.streamData {
		if err := stream.Send(message); err != nil {
			return err
		}
	}

	return nil
}

// SayHelloCS is client streaming handler
func (s *Greeter) SayHelloCS(stream Greeter_SayHelloCSServer) error {
	s.mutex.Lock()
	s.CallCounts["cs"]++
	s.mutex.Unlock()
	return errors.New("not implemented")
}

// SayHelloBidi duplex call handler
func (s *Greeter) SayHelloBidi(stream Greeter_SayHelloBidiServer) error {
	s.mutex.Lock()
	s.CallCounts["bidi"]++
	s.mutex.Unlock()
	return errors.New("not implemented")
}

// ResetCounters resets the call counts
func (s *Greeter) ResetCounters() {
	s.mutex.Lock()
	s.CallCounts["unary"] = 0
	s.CallCounts["ss"] = 0
	s.CallCounts["cs"] = 0
	s.CallCounts["bidi"] = 0
	s.mutex.Unlock()
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

	return &Greeter{streamData: streamData, CallCounts: m, mutex: &sync.Mutex{}}
}
