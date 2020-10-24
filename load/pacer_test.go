package load

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstantPacer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		freq    uint64
		max     uint64
		elapsed time.Duration
		hits    uint64
		wait    time.Duration
		stop    bool
	}{
		// 1 hit/sec, 0 hits sent, 0s elapsed => 1s until next hit
		{
			freq:    1,
			elapsed: 0,
			hits:    0,
			wait:    1000000000,
			stop:    false,
		},
		// 1 hit/sec, 0 hits sent, 0.1s elapsed => 0.9s until next hit
		{
			freq:    1,
			elapsed: 100 * time.Millisecond,
			hits:    0,
			wait:    900 * time.Millisecond,
			stop:    false,
		},
		// 1 hit/sec, 0 hits sent, 1s elapsed => 0s until next hit
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    0,
			wait:    0,
			stop:    false,
		},
		// 1 hit/sec, 0 hits sent, 2s elapsed => 0s (-1s) until next hit
		{
			freq:    1,
			elapsed: 2 * time.Second,
			hits:    0,
			wait:    0,
			stop:    false,
		},
		// 1 hit/sec, 1 hit sent, 1s elapsed => 1s until next hit
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    1,
			wait:    1 * time.Second,
			stop:    false,
		},
		// 1 hit/sec, 2 hits sent, 1s elapsed => 2s until next hit
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    2,
			wait:    2 * time.Second,
			stop:    false,
		},
		// 1 hit/sec, 10 hits sent, 1s elapsed => 10s until next hit
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    10,
			wait:    10 * time.Second,
			stop:    false,
		},
		// 1 hit/sec, 10 hits sent, 11s elapsed => 0s until next hit
		{
			freq:    1,
			elapsed: 11 * time.Second,
			hits:    10,
			wait:    0,
			stop:    false,
		},
		// 2 hit/sec, 9 hits sent, 4.9s elapsed => 100ms until next hit
		{
			freq:    2,
			elapsed: 4900 * time.Millisecond,
			hits:    9,
			wait:    100 * time.Millisecond,
			stop:    false,
		},
		// BAD TESTS
		// Zero frequency.
		{
			freq:    0,
			elapsed: 0,
			hits:    0,
			wait:    0,
			stop:    false,
		},
		// Large hits, overflow int64.
		{
			freq:    1,
			elapsed: time.Duration(math.MaxInt64),
			hits:    2562048,
			wait:    0,
			stop:    false,
		},
		// Max
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    10,
			wait:    10 * time.Second,
			stop:    false,
			max:     0,
		},
		{
			freq:    1,
			elapsed: 1 * time.Second,
			hits:    10,
			wait:    0,
			stop:    true,
			max:     7,
		},
	} {
		cp := ConstantPacer{Freq: tc.freq, Max: tc.max}
		wait, stop := cp.Pace(tc.elapsed, tc.hits)
		assert.Equal(t, tc.wait, wait)
		assert.Equal(t, tc.stop, stop)
	}
}

func TestConstantPacer_Rate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		freq    uint64
		elapsed time.Duration
		rate    float64
	}{
		{
			freq:    60,
			elapsed: 0,
			rate:    60,
		},
		{
			freq:    500,
			elapsed: 5 * time.Second,
			rate:    500.0,
		},
	} {
		cp := ConstantPacer{Freq: tc.freq}
		actual, expected := cp.Rate(tc.elapsed), tc.rate
		assert.True(t, floatEqual(actual, expected), "%s.Rate(_): actual %f, expected %f", cp, actual, expected)
	}
}

func TestConstantPacer_String(t *testing.T) {
	cp := ConstantPacer{Freq: 5}
	actual := cp.String()
	assert.Equal(t, "Constant{5 hits / 1s}", actual)
}

func TestLinearPacer(t *testing.T) {

	t.Parallel()

	for ti, tt := range []struct {
		// pacer config
		start        uint64
		slope        int64
		stopDuration time.Duration
		stopRate     uint64
		// params
		elapsed time.Duration
		hits    uint64
		// expected
		wait time.Duration
		stop bool
	}{
		// slope: 1, start 1
		{
			start:   1,
			slope:   1,
			elapsed: 0,
			hits:    0,
			wait:    1 * time.Second,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 1 * time.Second,
			hits:    0,
			wait:    0 * time.Millisecond,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 0,
			hits:    1,
			wait:    2 * time.Second,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 1 * time.Second,
			hits:    1,
			wait:    500 * time.Millisecond,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 1 * time.Second,
			hits:    2,
			wait:    1000 * time.Millisecond,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 2 * time.Second,
			hits:    2,
			wait:    0 * time.Millisecond,
			stop:    false,
		},
		{
			start:   1,
			slope:   1,
			elapsed: 2500 * time.Millisecond,
			hits:    5,
			wait:    500 * time.Millisecond,
			stop:    false,
		},
		// slope: 1, start 5
		{
			start:   5,
			slope:   1,
			elapsed: 0,
			hits:    0,
			wait:    200 * time.Millisecond,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 1 * time.Second,
			hits:    5,
			wait:    166666 * time.Microsecond,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 1200 * time.Millisecond,
			hits:    5,
			wait:    0 * time.Microsecond,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 2000 * time.Millisecond,
			hits:    6,
			wait:    0,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 2000 * time.Millisecond,
			hits:    7,
			wait:    0,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 2000 * time.Millisecond,
			hits:    11,
			wait:    142857 * time.Microsecond,
			stop:    false,
		},
		{
			start:   5,
			slope:   1,
			elapsed: 2000 * time.Millisecond,
			hits:    12,
			wait:    285714 * time.Microsecond,
			stop:    false,
		},
		// // slope: -1, start 20, various elapsed and hits
		{
			start:   20,
			slope:   -1,
			elapsed: 0,
			hits:    0,
			wait:    50 * time.Millisecond,
			stop:    false,
		},
		{
			start:   20,
			slope:   -1,
			elapsed: 1100 * time.Millisecond,
			hits:    0,
			wait:    0,
			stop:    false,
		},
		{
			start:   20,
			slope:   -1,
			elapsed: 50 * time.Millisecond,
			hits:    1,
			wait:    50 * time.Millisecond,
			stop:    false,
		},
		{
			start:   20,
			slope:   -1,
			elapsed: 50 * time.Millisecond,
			hits:    19,
			wait:    950 * time.Millisecond,
			stop:    false,
		},
		{
			start:   20,
			slope:   -1,
			elapsed: 950 * time.Millisecond,
			hits:    19,
			wait:    50 * time.Millisecond,
			stop:    false,
		},
		// slope: 1, stop rate
		{
			start:    1,
			slope:    1,
			elapsed:  0,
			stopRate: 20,
			hits:     0,
			wait:     1 * time.Second,
			stop:     false,
		},
		{
			start:    1,
			slope:    1,
			stopRate: 5,
			elapsed:  5 * time.Second,
			hits:     0,
			wait:     0,
			stop:     false,
		},
		{
			start:    1,
			slope:    1,
			stopRate: 5,
			elapsed:  5000 * time.Millisecond,
			hits:     17,
			wait:     600 * time.Millisecond,
			stop:     false,
		},
		// slope: 1, stop duration
		{
			start:        1,
			slope:        1,
			elapsed:      0,
			stopDuration: 5 * time.Second,
			hits:         0,
			wait:         1 * time.Second,
			stop:         false,
		},
		{
			start:        1,
			slope:        1,
			stopDuration: 5 * time.Second,
			elapsed:      2 * time.Second,
			hits:         0,
			wait:         0,
			stop:         false,
		},
		{
			start:        1,
			slope:        1,
			stopDuration: 5 * time.Second,
			elapsed:      2000 * time.Millisecond,
			hits:         5,
			wait:         1 * time.Second,
			stop:         false,
		},
		{
			start:        1,
			slope:        1,
			stopDuration: 5 * time.Second,
			elapsed:      5200 * time.Millisecond,
			hits:         18,
			wait:         466666 * time.Microsecond,
			stop:         false,
		},
	} {
		t.Run(strconv.Itoa(ti), func(t *testing.T) {
			p := LinearPacer{
				Start:        ConstantPacer{Freq: tt.start},
				Slope:        tt.slope,
				Stop:         ConstantPacer{Freq: tt.stopRate},
				LoadDuration: tt.stopDuration,
			}

			wait, stop := p.Pace(tt.elapsed, tt.hits)

			assert.True(t, durationEqual(tt.wait, wait),
				"%d: %+v.Pace(%s, %d) = (%s, %t); expected (%s, %t)", ti, &p, tt.elapsed, tt.hits, wait, stop, tt.wait, tt.stop)

			assert.Equal(t, tt.stop, stop)
		})
	}
}

func TestStepPacer_hits(t *testing.T) {
	t.Parallel()

	// TODO improve this to have different pacer params
	p := StepPacer{
		Start:        ConstantPacer{Freq: 10},
		StepDuration: 4 * time.Second,
		Step:         10,
	}

	for _, tc := range []struct {
		elapsed time.Duration
		hits    float64
	}{
		{0, 0},
		{1 * time.Second, 10},
		{2 * time.Second, 20},
		{6 * time.Second, 80},
	} {
		actual := p.hits(tc.elapsed)
		expected := tc.hits

		assert.True(t, floatEqual(actual, expected), "%+v.hits(%v) = %v, expected: %v", p, tc.elapsed, actual, expected)
	}
}

func TestStepPacer_Rate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		// pacer config
		start        uint64
		step         int64
		stepDuration time.Duration
		stop         uint64
		stopDuration time.Duration
		// params
		elapsed time.Duration
		// expected
		rate float64
	}{
		// step: 5, start: 1
		{
			start:        1,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      0,
			rate:         1,
		},
		{
			start:        1,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      1 * time.Second,
			rate:         1,
		},
		{
			start:        1,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      3 * time.Second,
			rate:         1,
		},
		{
			start:        1,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      4 * time.Second,
			rate:         6,
		},
		{
			start:        1,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      5 * time.Second,
			rate:         6,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         25,
			stopDuration: 0,
			elapsed:      9 * time.Second,
			rate:         15,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         25,
			stopDuration: 0,
			elapsed:      12 * time.Second,
			rate:         20,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         25,
			stopDuration: 0,
			elapsed:      22 * time.Second,
			rate:         25,
		},
		// start: 5, step: 5, stop duration: 25s
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      0,
			rate:         5,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      19 * time.Second,
			rate:         25,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      20 * time.Second,
			rate:         30,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      21 * time.Second,
			rate:         30,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      26 * time.Second,
			rate:         35,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 25 * time.Second,
			elapsed:      31 * time.Second,
			rate:         35,
		},
		// start: 15, step -5
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      0,
			rate:         15,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      3 * time.Second,
			rate:         15,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      4 * time.Second,
			rate:         10,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      5 * time.Second,
			rate:         10,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      11 * time.Second,
			rate:         5,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      12 * time.Second,
			rate:         0,
		},
		{
			start:        15,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      17 * time.Second,
			rate:         0,
		},
		// start: 20, step: -5, stop: 5
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         5,
			stopDuration: 0,
			elapsed:      0,
			rate:         20,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         5,
			stopDuration: 0,
			elapsed:      11 * time.Second,
			rate:         10,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         5,
			stopDuration: 0,
			elapsed:      12 * time.Second,
			rate:         5,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         5,
			stopDuration: 0,
			elapsed:      13 * time.Second,
			rate:         5,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         5,
			stopDuration: 0,
			elapsed:      22 * time.Second,
			rate:         5,
		},
		// start: 20, step: -5, stop: 10s
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      0,
			rate:         20,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      9 * time.Second,
			rate:         10,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      10 * time.Second,
			rate:         10,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      11 * time.Second,
			rate:         10,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      15 * time.Second,
			rate:         10,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 4 * time.Second,
			stop:         0,
			stopDuration: 10 * time.Second,
			elapsed:      22 * time.Second,
			rate:         10,
		},
	} {
		p := StepPacer{
			Start: ConstantPacer{Freq: tc.start},
			Step:  tc.step, StepDuration: tc.stepDuration,
			LoadDuration: tc.stopDuration, Stop: ConstantPacer{Freq: tc.stop}}

		actual := p.Rate(tc.elapsed)
		expected := tc.rate

		assert.True(t, floatEqual(actual, expected), "%+v.Rate(%v) = %v, expected: %v", p, tc.elapsed, actual, expected)
	}
}

func TestStepPacer(t *testing.T) {
	t.Parallel()

	for ti, tc := range []struct {
		// pacer config
		start        uint64
		step         int64
		stepDuration time.Duration
		stop         uint64
		stopDuration time.Duration
		max          uint64
		// params
		elapsed time.Duration
		hits    uint64
		// expected
		wait       time.Duration
		stopResult bool
	}{
		// start: 5, step: 5
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0 * time.Second,
			elapsed:      0 * time.Second,
			hits:         0,
			wait:         200 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0 * time.Second,
			elapsed:      1 * time.Second,
			hits:         4,
			wait:         0 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0 * time.Second,
			elapsed:      1 * time.Second,
			hits:         6,
			wait:         400 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0 * time.Second,
			elapsed:      4200 * time.Millisecond,
			hits:         25,
			wait:         1 * time.Second,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0 * time.Second,
			elapsed:      5000 * time.Millisecond,
			hits:         25,
			wait:         100 * time.Millisecond,
			stopResult:   false,
		},
		// start: 5, step: 5, stop: 25
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			stopDuration: 0 * time.Second,
			elapsed:      5000 * time.Millisecond,
			hits:         25,
			wait:         100 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			stopDuration: 0 * time.Second,
			elapsed:      20 * time.Second,
			hits:         250,
			wait:         40 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			stopDuration: 0 * time.Second,
			elapsed:      30 * time.Second,
			hits:         450,
			wait:         0 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			stopDuration: 0 * time.Second,
			elapsed:      30 * time.Second,
			hits:         500,
			wait:         40 * time.Millisecond,
			stopResult:   false,
		},
		// start: 5, step: 5, stop: 20s
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      5000 * time.Millisecond,
			hits:         25,
			wait:         100 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      19 * time.Second,
			hits:         25,
			wait:         0 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      20 * time.Second,
			hits:         250,
			wait:         40 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      30 * time.Second,
			hits:         400,
			wait:         0 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      30 * time.Second,
			hits:         500,
			wait:         40 * time.Millisecond,
			stopResult:   false,
		},
		// start: 20, step: -5,
		{
			start:        20,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      0 * time.Millisecond,
			hits:         0,
			wait:         50 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      5000 * time.Millisecond,
			hits:         100,
			wait:         66666666 * time.Nanosecond,
			stopResult:   false,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      5000 * time.Millisecond,
			hits:         100,
			wait:         66666666 * time.Nanosecond,
			stopResult:   false,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      20 * time.Second,
			hits:         249,
			wait:         0,
			stopResult:   false,
		},
		{
			start:        20,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 0,
			elapsed:      20 * time.Second,
			hits:         250,
			wait:         0,
			stopResult:   true,
		},
		{
			start:        30,
			step:         -5,
			stepDuration: 5 * time.Second,
			stop:         0,
			stopDuration: 20 * time.Second,
			elapsed:      30 * time.Second,
			hits:         550,
			wait:         100 * time.Millisecond,
			stopResult:   false,
		},
		// Max
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			stopDuration: 0 * time.Second,
			max:          100,
			elapsed:      5000 * time.Millisecond,
			hits:         25,
			wait:         100 * time.Millisecond,
			stopResult:   false,
		},
		{
			start:        5,
			step:         5,
			stepDuration: 5 * time.Second,
			stop:         25,
			max:          10,
			stopDuration: 0 * time.Second,
			elapsed:      5000 * time.Millisecond,
			hits:         25,
			wait:         0,
			stopResult:   true,
		},
	} {
		t.Run(strconv.Itoa(ti), func(t *testing.T) {
			p := StepPacer{
				Start: ConstantPacer{Freq: tc.start, Max: tc.max},
				Max:   tc.max,
				Step:  tc.step, StepDuration: tc.stepDuration,
				LoadDuration: tc.stopDuration, Stop: ConstantPacer{Freq: tc.stop}}

			wait, stop := p.Pace(tc.elapsed, tc.hits)

			assert.Equal(t, tc.wait, wait, "%+v.Pace(%v, %v) = %v, expected: %v", p, tc.elapsed, tc.hits, wait, tc.wait)
			assert.Equal(t, tc.stopResult, stop, "%+v.Pace(%v, %v) = %v, expected: %v", p, tc.elapsed, tc.hits, stop, tc.stopResult)
		})
	}
}

func TestStepPacer_String(t *testing.T) {
	p := StepPacer{
		Start: ConstantPacer{Freq: 5, Max: 100},
		Max:   100,
		Step:  2, StepDuration: 5 * time.Second,
		LoadDuration: 25 * time.Second, Stop: ConstantPacer{Freq: 25}}

	actual := p.String()
	assert.Equal(t, "Step{Step: 2 hits / 5s}", actual)
}

// Stolen from https://github.com/google/go-cmp/cmp/cmpopts/equate.go
// to avoid an unwieldy dependency. Both fraction and margin set at 1e-6.
func floatEqual(x, y float64) bool {
	relMarg := 1e-6 * math.Min(math.Abs(x), math.Abs(y))
	return math.Abs(x-y) <= math.Max(1e-6, relMarg)
}

// A similar function to the above because SinePacer.Pace has discrete
// inputs and outputs but uses floats internally, and sometimes the
// floating point imprecision leaks out :-(
func durationEqual(x, y time.Duration) bool {
	diff := x - y
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Microsecond
}
