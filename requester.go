package main

import (
	"sync"
	"time"
)

// Max size of the buffer of result channel.
const maxResult = 1000000
const maxIdleConn = 500

type Result struct {
	err        error
	statusCode int
	duration   time.Duration
}

// Requestor is used for doing the requests
type Requestor struct {
	config  *Config
	results chan *Result
	stopCh  chan struct{}
	start   time.Time
}

// New creates new Requestor
func New(c *Config) *Requestor {
	return &Requestor{config: c}
}

// Run makes all the requests, prints the summary.
// It blocks until all work is done.
func (b *Requestor) Run() {
	b.results = make(chan *Result, min(b.config.C*1000, maxResult))
	b.stopCh = make(chan struct{}, b.config.C)
	b.start = time.Now()
	// b.report = newReport(b.writer(), b.results, b.Output, b.N)
	// // Run the reporter first, it polls the result channel until it is closed.
	// go func() {
	// 	runReporter(b.report)
	// }()
	b.runWorkers()
	b.Finish()
}

func (b *Requestor) Stop() {
	// Send stop signal so that workers can stop gracefully.
	for i := 0; i < b.config.C; i++ {
		b.stopCh <- struct{}{}
	}
}

func (b *Requestor) Finish() {
	close(b.results)
	total := time.Now().Sub(b.start)
}

func (b *Requestor) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(b.config.C)

	// TODO create connection

	// Ignore the case where b.N % b.C != 0.
	for i := 0; i < b.config.C; i++ {
		go func() {
			// b.runWorker(client, b.config.N/b.config.C)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (b *Requestor) runWorker( /* client, */ n int) {
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
			// b.makeRequest(client)
			b.makeRequest()
		}
	}
}

func (b *Requestor) makeRequest( /*c *http.Client*/ ) {
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
