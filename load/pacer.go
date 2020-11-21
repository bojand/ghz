package load

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// nano is the const for number of nanoseconds in a second
const nano = 1e9

// Pacer defines the interface to control the rate of hit.
type Pacer interface {
	// Pace returns the duration the attacker should wait until
	// making next hit, given an already elapsed duration and
	// completed hits. If the second return value is true, an attacker
	// should stop sending hits.
	Pace(elapsed time.Duration, hits uint64) (wait time.Duration, stop bool)

	// Rate returns a Pacer's instantaneous hit rate (per seconds)
	// at the given elapsed duration of an attack.
	Rate(elapsed time.Duration) float64
}

// A ConstantPacer defines a constant rate of hits.
type ConstantPacer struct {
	Freq uint64 // Frequency of hits per second
	Max  uint64 // Optional maximum allowed hits
}

// String returns a pretty-printed description of the ConstantPacer's behaviour:
//   ConstantPacer{Freq: 1} => Constant{1 hits / 1s}
func (cp *ConstantPacer) String() string {
	return fmt.Sprintf("Constant{%d hits / 1s}", cp.Freq)
}

// Pace determines the length of time to sleep until the next hit is sent.
func (cp *ConstantPacer) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {

	if cp.Max > 0 && hits >= cp.Max {
		return 0, true
	}

	if cp.Freq == 0 {
		return 0, false // Zero value = infinite rate
	}

	expectedHits := uint64(cp.Freq) * uint64(elapsed/nano)
	if hits < expectedHits {
		// Running behind, send next hit immediately.
		return 0, false
	}

	interval := uint64(nano / int64(cp.Freq))
	if math.MaxInt64/interval < hits {
		// We would overflow delta if we continued, so stop the attack.
		return 0, true
	}

	delta := time.Duration((hits + 1) * interval)
	// Zero or negative durations cause time.Sleep to return immediately.
	return delta - elapsed, false
}

// Rate returns a ConstantPacer's instantaneous hit rate (i.e. requests per second)
// at the given elapsed duration of an attack. Since it's constant, the return
// value is independent of the given elapsed duration.
func (cp *ConstantPacer) Rate(elapsed time.Duration) float64 {
	return cp.hitsPerNs() * 1e9
}

// hitsPerNs returns the rate in fractional hits per nanosecond.
func (cp *ConstantPacer) hitsPerNs() float64 {
	return float64(cp.Freq) / nano
}

// StepPacer paces an attack by starting at a given request rate
// and increasing or decreasing with steps at a given step interval and duration.
type StepPacer struct {
	Start        ConstantPacer // Constant start rate
	Step         int64         // Step value
	StepDuration time.Duration // Step duration
	Stop         ConstantPacer // Optional constant stop value
	LoadDuration time.Duration // Optional maximum load duration
	Max          uint64        // Optional maximum allowed hits

	once     sync.Once
	init     bool // TOOO improve this
	constAt  time.Duration
	baseHits uint64
}

func (p *StepPacer) initialize() {

	if p.StepDuration == 0 {
		panic("StepPacer.StepDuration cannot be 0")
	}

	if p.Step == 0 {
		panic("StepPacer.Step cannot be 0")
	}

	if p.Start.Freq == 0 {
		panic("Start.Freq cannot be 0")
	}

	if p.init {
		return
	}

	p.init = true

	if p.LoadDuration > 0 {
		p.constAt = p.LoadDuration

		if p.Stop.Freq == 0 {
			steps := p.constAt.Nanoseconds() / p.StepDuration.Nanoseconds()

			p.Stop.Freq = p.Start.Freq + uint64(int64(p.Step)*steps)
		}
	} else if p.Stop.Freq > 0 && p.constAt == 0 {
		stopRPS := float64(p.Stop.Freq)

		if p.Step > 0 {
			t := time.Duration(0)
			for {
				if p.Rate(t) > stopRPS {
					p.constAt = t
					break
				}
				t = t + p.StepDuration
			}
		} else {
			t := time.Duration(0)
			for {
				if p.Rate(t) < stopRPS {
					p.constAt = t
					break
				}
				t = t + p.StepDuration
			}
		}
	}

	if p.constAt > 0 {
		p.baseHits = uint64(p.hits(p.constAt))
	}
}

// Pace determines the length of time to sleep until the next hit is sent.
func (p *StepPacer) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {

	if p.Max > 0 && hits >= p.Max {
		return 0, true
	}

	p.once.Do(p.initialize)

	expectedHits := p.hits(elapsed)

	if hits < uint64(expectedHits) {
		// Running behind, send next hit immediately.
		return 0, false
	}

	// const part
	if p.constAt > 0 && elapsed >= p.constAt {
		if p.Stop.Freq == 0 {
			return 0, true
		}

		return p.Stop.Pace(elapsed-p.constAt, hits-p.baseHits)
	}

	rate := p.Rate(elapsed)
	interval := nano / rate

	if n := uint64(interval); n != 0 && math.MaxInt64/n < hits {
		// We would overflow wait if we continued, so stop the attack.
		return 0, true
	}

	delta := float64(hits+1) - expectedHits
	wait := time.Duration(interval * delta)

	// if wait > nano {
	// 	intervals := elapsed / nano
	// 	wait = (intervals+1)*nano - elapsed
	// }

	return wait, false
}

// Rate returns a StepPacer's instantaneous hit rate (i.e. requests per second)
// at the given elapsed duration.
func (p *StepPacer) Rate(elapsed time.Duration) float64 {
	p.initialize()

	t := elapsed

	if p.constAt > 0 && elapsed >= p.constAt {
		return float64(p.Stop.Freq)
	}

	steps := t.Nanoseconds() / p.StepDuration.Nanoseconds()

	rate := (p.Start.hitsPerNs() + float64(int64(p.Step)*steps)/nano) * 1e9

	if rate < 0 {
		rate = 0
	}

	return rate
}

// hits returns the number of hits that have been sent at elapsed duration t.
func (p *StepPacer) hits(t time.Duration) float64 {
	if t < 0 {
		return 0
	}

	steps := t.Nanoseconds() / p.StepDuration.Nanoseconds()

	base := p.Start.hitsPerNs() * 1e9

	// first step
	var s float64
	if steps > 0 {
		s = p.StepDuration.Seconds() * base
	} else {
		s = t.Seconds() * base
	}

	// previous steps: 1...n
	for i := int64(1); i < steps; i++ {
		d := time.Duration(p.StepDuration.Nanoseconds() * i)
		r := p.Rate(d)
		ch := r * p.StepDuration.Seconds()
		s = s + ch
	}

	c := float64(0)
	if steps > 0 {
		// current step
		elapsed := time.Duration(t.Nanoseconds() - steps*p.StepDuration.Nanoseconds())
		c = elapsed.Seconds() * p.Rate(t)
	}

	return s + c
}

// String returns a pretty-printed description of the StepPacer's behaviour:
//   StepPacer{Step: 1, StepDuration: 5s} => Step{Step:1 hits / 5s}
func (p *StepPacer) String() string {
	return fmt.Sprintf("Step{Step: %d hits / %s}", p.Step, p.StepDuration.String())
}

// LinearPacer paces the hit rate by starting at a given request rate
// and increasing linearly with the given slope at 1s interval.
type LinearPacer struct {
	Start        ConstantPacer // Constant start rate
	Slope        int64         // Slope value to change the rate
	Stop         ConstantPacer // Constant stop rate
	LoadDuration time.Duration // Total maximum load duration
	Max          uint64        // Maximum number of hits

	once sync.Once
	sp   StepPacer
}

// initializes the wrapped step pacer
func (p *LinearPacer) initialize() {
	if p.Start.Freq == 0 {
		panic("LinearPacer.Start cannot be 0")
	}

	if p.Slope == 0 {
		panic("LinearPacer.Slope cannot be 0")
	}

	p.once.Do(func() {
		p.sp = StepPacer{
			Start:        p.Start,
			Step:         p.Slope,
			StepDuration: time.Second,
			Stop:         p.Stop,
			LoadDuration: p.LoadDuration,
		}

		p.sp.initialize()
	})
}

// Pace determines the length of time to sleep until the next hit is sent.
func (p *LinearPacer) Pace(elapsed time.Duration, hits uint64) (time.Duration, bool) {
	if p.Max > 0 && hits >= p.Max {
		return 0, true
	}

	p.initialize()

	return p.sp.Pace(elapsed, hits)
}

// Rate returns a LinearPacer's instantaneous hit rate (i.e. requests per second)
// at the given elapsed duration.
func (p *LinearPacer) Rate(elapsed time.Duration) float64 {

	p.initialize()

	return p.sp.Rate(elapsed)
}

// String returns a pretty-printed description of the LinearPacer's behaviour:
//   LinearPacer{Slope: 1} => Linear{1 hits / 1s}
func (p *LinearPacer) String() string {
	return fmt.Sprintf("Linear{%d hits / 1s}", p.Slope)
}
