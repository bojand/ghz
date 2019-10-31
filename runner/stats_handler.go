package runner

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

// StatsHandler is for gRPC stats
type statsHandler struct {
	results chan *callResult

	lock   sync.RWMutex
	ignore bool
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
	switch rs := rs.(type) {
	case *stats.End:
		ign := false
		c.lock.RLock()
		ign = c.ignore
		c.lock.RUnlock()

		if !ign {
			rpcStats := rs
			end := time.Now()
			duration := end.Sub(rpcStats.BeginTime)

			var st string
			s, ok := status.FromError(rpcStats.Error)
			if ok {
				st = s.Code().String()
			}

			c.results <- &callResult{rpcStats.Error, st, duration, end}
		}
	}
}

func (c *statsHandler) Ignore(val bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.ignore = val
}

// TagRPC implements per-RPC context management.
func (c *statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}
