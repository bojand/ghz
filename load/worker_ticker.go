package load

import (
	"fmt"
	"time"
)

// WorkerTicker is the interface for worker ticker which controls worker parallelism
type WorkerTicker interface {
	Ticker() <-chan TickValue
	Run()
	Finish()
}

// TickValue is the delta value
type TickValue struct {
	Delta int
}

// ConstWorkerTicker is the const worker
type ConstWorkerTicker struct {
	C chan TickValue
	N uint
}

// Ticker returns the ticker channer
func (c *ConstWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker
func (c *ConstWorkerTicker) Run() {
	c.C <- TickValue{Delta: int(c.N)}
}

// Finish stops
func (c *ConstWorkerTicker) Finish() {
	close(c.C)
}

// StepWorkerTicker is the worker ticker that implements step adjustments to concurrency
type StepWorkerTicker struct {
	C chan TickValue

	Start        uint
	Step         int
	StepDuration time.Duration
	Stop         uint
	LoadDuration time.Duration
}

// Ticker returns the ticker channer
func (c *StepWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker
func (c *StepWorkerTicker) Run() {

	stepUp := c.Step > 0
	wc := int(c.Start)
	done := make(chan bool)

	ticker := time.NewTicker(c.StepDuration)
	defer ticker.Stop()

	begin := time.Now()

	c.C <- TickValue{Delta: int(c.Start)}

	go func() {
		fmt.Println("step ticker start")
		for range ticker.C {
			fmt.Println("worker ticker", time.Since(begin))
			if c.LoadDuration > 0 && time.Since(begin) > c.LoadDuration {
				fmt.Println("duration stop reached", wc, time.Since(begin))

				if c.Stop > 0 {
					c.C <- TickValue{Delta: int(c.Stop - uint(wc))}
				}

				done <- true
				return
			} else if (c.Stop > 0 && stepUp && wc >= int(c.Stop)) ||
				(!stepUp && wc <= int(c.Stop)) || wc <= 0 {
				fmt.Println("stop reached", wc, time.Since(begin))
				done <- true
				return
			} else {
				c.C <- TickValue{Delta: c.Step}
				wc = wc + c.Step
			}
		}
	}()

	<-done
}

// Finish stops
func (c *StepWorkerTicker) Finish() {
	close(c.C)
}

// LineWorkerTicker is the worker ticker that implements line adjustments to concurrency
type LineWorkerTicker struct {
	C chan TickValue

	Start        uint
	Slope        int
	Stop         uint
	LoadDuration time.Duration

	stepTicker StepWorkerTicker
}

// Ticker returns the ticker channer
func (c *LineWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker
func (c *LineWorkerTicker) Run() {

	c.stepTicker = StepWorkerTicker{
		C:            c.C,
		Start:        c.Start,
		Step:         c.Slope,
		StepDuration: 1 * time.Second,
		Stop:         c.Stop,
		LoadDuration: c.LoadDuration,
	}

	c.stepTicker.Run()
}

// Finish stops
func (c *LineWorkerTicker) Finish() {
	c.stepTicker.Finish()
}
