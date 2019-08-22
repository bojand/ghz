package runner

import (
	"context"
	"time"

	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

// StatsHandler is for gRPC stats
type statsHandler struct {
	results chan *callResult
}

// HandleConn handle the connection
func (c *statsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (c *statsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (c *statsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	switch rs.(type) {
	case *stats.End:
		rpcStats := rs.(*stats.End)
		end := time.Now()
		duration := end.Sub(rpcStats.BeginTime)

		var st string
		s, ok := status.FromError(rpcStats.Error)
		if ok {
			st = s.Code().String()
		}

		c.results <- &callResult{rpcStats.Error, st, duration, rpcStats.BeginTime}
	}
}

// TagRPC implements per-RPC context management.
func (c *statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}
