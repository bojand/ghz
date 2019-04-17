package helloworld

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc/stats"
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

	Stats *HWStatsHandler

	mutex      *sync.RWMutex
	callCounts map[CallType]int
}

func RandomSleep() {
	msCount := rand.Intn(4) + 1
	time.Sleep(time.Millisecond * time.Duration(msCount))
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	s.mutex.Lock()
	s.callCounts[Unary]++
	s.mutex.Unlock()

	RandomSleep()

	return &HelloReply{Message: "Hello " + in.Name}, nil
}

// SayHellos lists all hellos
func (s *Greeter) SayHellos(req *HelloRequest, stream Greeter_SayHellosServer) error {
	s.mutex.Lock()
	s.callCounts[ServerStream]++
	s.mutex.Unlock()

	RandomSleep()

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

	RandomSleep()

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

	RandomSleep()

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

	if s.Stats != nil {
		s.Stats.mutex.Lock()
		s.Stats.connCount = 0
		s.Stats.mutex.Unlock()
	}
}

// GetCount gets the count for specific call type
func (s *Greeter) GetCount(key CallType) int {
	s.mutex.RLock()
	val, ok := s.callCounts[key]
	s.mutex.RUnlock()
	if ok {
		return val
	}
	return -1
}

// GetConnectionCount gets the connection count
func (s *Greeter) GetConnectionCount() int {
	return s.Stats.GetConnectionCount()
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

	return &Greeter{streamData: streamData, callCounts: m, mutex: &sync.RWMutex{}}
}

// NewHWStats creates new stats handler
func NewHWStats() *HWStatsHandler {
	return &HWStatsHandler{connCount: 0, mutex: &sync.RWMutex{}}
}

// HWStatsHandler is for gRPC stats
type HWStatsHandler struct {
	mutex     *sync.RWMutex
	connCount int
}

// GetConnectionCount gets the connection count
func (c *HWStatsHandler) GetConnectionCount() int {
	c.mutex.RLock()
	val := c.connCount
	c.mutex.RUnlock()

	return val
}

// HandleConn handle the connection
func (c *HWStatsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	fmt.Println("!!! HandleConn")
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *HWStatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	fmt.Println("!!! TagConn")

	c.mutex.Lock()
	c.connCount++
	c.mutex.Unlock()

	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (c *HWStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	// no-op
}

// TagRPC implements per-RPC context management.
func (c *HWStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}