package helloworld

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

// CallType represents one of the gRPC call types:
// unary, client streaming, server streaming, bidi
type CallType string

// Unary is a unary call
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
	calls      map[CallType][][]*HelloRequest
}

func randomSleep() {
	msCount := rand.Intn(4) + 1
	time.Sleep(time.Millisecond * time.Duration(msCount))
}

func (s *Greeter) recordCall(ct CallType) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.callCounts[ct]++
	var messages []*HelloRequest
	s.calls[ct] = append(s.calls[ct], messages)

	return len(s.calls[ct]) - 1
}

func (s *Greeter) recordMessage(ct CallType, callIdx int, msg *HelloRequest) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.calls[ct][callIdx] = append(s.calls[ct][callIdx], msg)
}

// SayHello implements helloworld.GreeterServer
func (s *Greeter) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	callIdx := s.recordCall(Unary)
	s.recordMessage(Unary, callIdx, in)

	randomSleep()

	return &HelloReply{Message: "Hello " + in.Name}, nil
}

// SayHellos lists all hellos
func (s *Greeter) SayHellos(req *HelloRequest, stream Greeter_SayHellosServer) error {
	callIdx := s.recordCall(ServerStream)
	s.recordMessage(ServerStream, callIdx, req)

	randomSleep()

	for _, msg := range s.streamData {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

// SayHelloCS is client streaming handler
func (s *Greeter) SayHelloCS(stream Greeter_SayHelloCSServer) error {
	callIdx := s.recordCall(ClientStream)

	randomSleep()

	msgCount := 0

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			msgStr := fmt.Sprintf("Hello count: %d", msgCount)
			return stream.SendAndClose(&HelloReply{Message: msgStr})
		}
		if err != nil {
			return err
		}
		s.recordMessage(ClientStream, callIdx, in)
		msgCount++
	}
}

// SayHelloBidi duplex call handler
func (s *Greeter) SayHelloBidi(stream Greeter_SayHelloBidiServer) error {
	callIdx := s.recordCall(Bidi)

	randomSleep()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		s.recordMessage(Bidi, callIdx, in)
		msg := "Hello " + in.Name
		if err := stream.Send(&HelloReply{Message: msg}); err != nil {
			return err
		}
	}
}

// ResetCounters resets the call counts
func (s *Greeter) ResetCounters() {
	s.mutex.Lock()

	s.callCounts = make(map[CallType]int)
	s.callCounts[Unary] = 0
	s.callCounts[ServerStream] = 0
	s.callCounts[ClientStream] = 0
	s.callCounts[Bidi] = 0

	s.calls = make(map[CallType][][]*HelloRequest)
	s.calls[Unary] = make([][]*HelloRequest, 0)
	s.calls[ServerStream] = make([][]*HelloRequest, 0)
	s.calls[ClientStream] = make([][]*HelloRequest, 0)
	s.calls[Bidi] = make([][]*HelloRequest, 0)

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

// GetCalls gets the received messages for specific call type
func (s *Greeter) GetCalls(key CallType) [][]*HelloRequest {
	s.mutex.Lock()
	val, ok := s.calls[key]
	s.mutex.Unlock()

	if ok {
		return val
	}
	return nil
}

// GetConnectionCount gets the connection count
func (s *Greeter) GetConnectionCount() int {
	return s.Stats.GetConnectionCount()
}

// NewGreeter creates new greeter server
func NewGreeter() *Greeter {
	streamData := []*HelloReply{
		{Message: "Hello Bob"},
		{Message: "Hello Kate"},
		{Message: "Hello Jim"},
		{Message: "Hello Sara"},
	}

	greeter := &Greeter{streamData: streamData, mutex: &sync.RWMutex{}}
	greeter.ResetCounters()

	return greeter
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

// GetCountByWorker gets count of requests by goroutine
func (s *Greeter) GetCountByWorker(key CallType) map[string]int {
	s.mutex.Lock()
	val, ok := s.calls[key]
	s.mutex.Unlock()

	if !ok {
		return nil
	}

	counts := make(map[string]int)

	for _, reqs := range val {
		for _, req := range reqs {
			name := req.GetName()
			if strings.Contains(name, "worker:") {
				parts := strings.Split(name, ":")
				wid := parts[len(parts)-1]
				wc, ok := counts[wid]
				if !ok {
					counts[wid] = 0
				}

				counts[wid] = wc + 1
			}
		}
	}

	return counts
}

// HandleConn handle the connection
func (c *HWStatsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *HWStatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
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
