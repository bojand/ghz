package load

import (
	"time"
)

// WorkerTicker is the interface controlling worker parallelism.
type WorkerTicker interface {
	// Ticker returns a channel which sends TickValues
	// When a value is received the number of workers should be appropriately
	// increased or decreased given by the delta property.
	Ticker() <-chan TickValue

	// Run starts the worker ticker
	Run()

	// Finish closes the channel
	Finish()
}

// TickValue is the tick value sent over the ticker channel.
type TickValue struct {
	Delta int  // Delta value representing worker increase or decrease
	Done  bool // A flag representing whether the ticker is done running. Once true no more values should be received over the ticker channel.
}

// ConstWorkerTicker represents a constant number of workers.
// It would send one value for initial number of workers to start.
type ConstWorkerTicker struct {
	C chan TickValue // The tick value channel
	N uint           // The number of workers
}

// Ticker returns the ticker channel.
func (c *ConstWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker.
func (c *ConstWorkerTicker) Run() {
	c.C <- TickValue{Delta: int(c.N), Done: true}
}

// Finish closes the channel.
func (c *ConstWorkerTicker) Finish() {
	close(c.C)
}

// StepWorkerTicker is the worker ticker that implements step adjustments to worker concurrency.
type StepWorkerTicker struct {
	C chan TickValue // The tick value channel

	Start        uint          // Starting number of workers
	Step         int           // Step change
	StepDuration time.Duration // Duration to apply the step change
	Stop         uint          // Final number of workers
	MaxDuration  time.Duration // Maximum duration
}

// Ticker returns the ticker channel.
func (c *StepWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker.
func (c *StepWorkerTicker) Run() {

	stepUp := c.Step > 0
	wc := int(c.Start)
	done := make(chan bool)

	ticker := time.NewTicker(c.StepDuration)
	defer ticker.Stop()

	begin := time.Now()

	c.C <- TickValue{Delta: int(c.Start)}

	go func() {
		for range ticker.C {
			// we have load duration and we eclipsed it
			if c.MaxDuration > 0 && time.Since(begin) >= c.MaxDuration {
				if stepUp && c.Stop > 0 && c.Stop >= uint(wc) {
					// if we have step up and stop value is > current count
					// send the final diff
					c.C <- TickValue{Delta: int(c.Stop - uint(wc)), Done: true}
				} else if !stepUp && c.Stop > 0 && c.Stop <= uint(wc) {
					// if we have step down and stop value is < current count
					// send the final diff
					c.C <- TickValue{Delta: int(c.Stop - uint(wc)), Done: true}
				} else {
					// send done signal
					c.C <- TickValue{Delta: 0, Done: true}
				}

				done <- true
				return
			} else if (c.MaxDuration == 0) && ((c.Stop > 0 && stepUp && wc >= int(c.Stop)) ||
				(!stepUp && wc <= int(c.Stop))) {
				// we do not have load duration
				// if we have stop and are step up and current count >= stop
				// or if we have stop and are step down and current count <= stop
				// send done signal

				c.C <- TickValue{Delta: 0, Done: true}
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

// Finish closes the channel.
func (c *StepWorkerTicker) Finish() {
	close(c.C)
}

// LineWorkerTicker is the worker ticker that implements line adjustments to concurrency.
// Essentially this is same as step worker with 1s step duration.
type LineWorkerTicker struct {
	C chan TickValue // The tick value channel

	Start       uint          // Starting number of workers
	Slope       int           // Slope value to adjust the number of workers
	Stop        uint          // Final number of workers
	MaxDuration time.Duration // Maximum adjustment duration

	stepTicker StepWorkerTicker
}

// Ticker returns the ticker channel.
func (c *LineWorkerTicker) Ticker() <-chan TickValue {
	return c.C
}

// Run runs the ticker.
func (c *LineWorkerTicker) Run() {

	c.stepTicker = StepWorkerTicker{
		C:            c.C,
		Start:        c.Start,
		Step:         c.Slope,
		StepDuration: 1 * time.Second,
		Stop:         c.Stop,
		MaxDuration:  c.MaxDuration,
	}

	c.stepTicker.Run()
}

// Finish closes the internal tick value channel.
func (c *LineWorkerTicker) Finish() {
	c.stepTicker.Finish()
}
