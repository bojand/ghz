package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

// Max size of the buffer of result channel.
const maxResult = 1000000
const maxIdleConn = 500

// Result is a result of a call
type Result struct {
	err      error
	status   string
	duration time.Duration
}

// Report holds the data for the full test
type Report struct {
	count    uint32
	duration time.Duration
	average  time.Duration
}

// Requestor is used for doing the requests
type Requestor struct {
	cc    *grpc.ClientConn
	stub  grpcdynamic.Stub
	mtd   *desc.MethodDescriptor
	input *dynamic.Message
	reqMD *metadata.MD

	config  *Config
	results chan *Result
	stopCh  chan bool
	start   time.Time

	totalCount uint32
	avgTotal   float64
	done       chan bool
}

// New creates new Requestor
func New(c *Config, mtd *desc.MethodDescriptor) (*Requestor, error) {
	input := dynamic.NewMessage(mtd.GetInputType())
	if input == nil {
		return nil, fmt.Errorf("No input type of method: %s", mtd.GetName())
	}

	// payload
	for k, v := range *c.Data {
		err := input.TrySetFieldByName(k, v)
		if err != nil {
			return nil, err
		}
	}

	// metadata
	var reqMD *metadata.MD
	if c.Metadata != nil && len(*c.Metadata) > 0 {
		md := metadata.New(*c.Metadata)
		reqMD = &md
	}

	return &Requestor{config: c, input: input, reqMD: reqMD, mtd: mtd}, nil
}

// Run makes all the requests, prints the summary.
// It blocks until all work is done.
func (b *Requestor) Run() (*Report, error) {
	b.results = make(chan *Result, min(b.config.C*1000, maxResult))
	b.stopCh = make(chan bool, b.config.C)
	b.start = time.Now()

	b.totalCount = 0
	b.avgTotal = 0
	b.done = make(chan bool, 1)

	cc, err := b.connect()
	if err != nil {
		return nil, err
	}

	b.cc = cc
	defer cc.Close()

	b.stub = grpcdynamic.NewStub(cc)

	go func() {
		for res := range b.results {
			// fmt.Printf("Result: %+v Duration: %+v\n", *res, res.duration)
			b.avgTotal += res.duration.Seconds()
			b.totalCount++
		}
		b.done <- true
	}()

	b.runWorkers()

	report := b.Finish()

	return report, nil
}

// Stop stops the test
func (b *Requestor) Stop() {
	// Send stop signal so that workers can stop gracefully.
	for i := 0; i < b.config.C; i++ {
		b.stopCh <- true
	}
}

// Finish finishes the test run
func (b *Requestor) Finish() *Report {
	close(b.results)
	total := time.Now().Sub(b.start)

	// Wait until the reporter is done.
	<-b.done
	average := b.avgTotal / float64(b.totalCount)

	avgDuration := time.Duration(average * float64(time.Second))

	return &Report{b.totalCount, total, avgDuration}
}

func (b *Requestor) connect() (*grpc.ClientConn, error) {
	credOptions, err := CreateClientCredOption(b.config)
	if err != nil {
		return nil, err
	}

	// create client connection
	return grpc.Dial(b.config.Host, grpc.WithStatsHandler(&StatsHandler{b.results}), credOptions)
}

func (b *Requestor) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.config.C)

	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.config.C; i++ {
		go func() {
			defer wg.Done()

			b.runWorker(b.config.N / b.config.C)
		}()
	}
	wg.Wait()
}

func (b *Requestor) runWorker(n int) {
	var throttle <-chan time.Time
	if b.config.QPS > 0 {
		throttle = time.Tick(time.Duration(1e6/(b.config.QPS)) * time.Microsecond)
	}

	for i := 0; i < n; i++ {
		// Check if application is stopped. Do not send into a closed channel.
		select {
		case <-b.stopCh:
			return
		default:
			if b.config.QPS > 0 {
				<-throttle
			}
			b.makeRequest()
		}
	}
}

func (b *Requestor) makeRequest() {
	// create call context
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := time.Duration(int64(b.config.Timeout) * int64(time.Second))
	ctx, _ = context.WithTimeout(ctx, timeout)

	// include the metadata
	if b.reqMD != nil {
		ctx = metadata.NewOutgoingContext(ctx, *b.reqMD)
	}

	if !b.mtd.IsClientStreaming() && !b.mtd.IsServerStreaming() {
		b.stub.InvokeRpc(ctx, b.mtd, b.input)
	}
}

// CreateClientCredOption creates the credential dial options based on config
func CreateClientCredOption(config *Config) (grpc.DialOption, error) {
	credOptions := grpc.WithInsecure()
	if strings.TrimSpace(config.Cert) != "" {
		creds, err := credentials.NewClientTLSFromFile(config.Cert, "")
		if err != nil {
			return nil, err
		}
		credOptions = grpc.WithTransportCredentials(creds)
	}

	return credOptions, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// StatsHandler is for gRPC stats
type StatsHandler struct {
	results chan *Result
}

// HandleConn handle the connection
func (c *StatsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *StatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (c *StatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	switch rs.(type) {
	case *stats.End:
		rpcStats := rs.(*stats.End)
		end := time.Now()
		duration := end.Sub(rpcStats.BeginTime)

		var st string
		if rpcStats.Error != nil {
			s, ok := status.FromError(rpcStats.Error)
			if ok {
				st = s.Code().String()
			}
		}

		c.results <- &Result{rpcStats.Error, st, duration}
	}
}

// TagRPC implements per-RPC context management.
func (c *StatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}
