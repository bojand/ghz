package load

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstWorkerTicker(t *testing.T) {
	wt := ConstWorkerTicker{N: 5, C: make(chan TickValue)}
	defer wt.Finish()

	wct := wt.Ticker()

	assert.NotNil(t, wct)

	go func() {
		wt.Run()
	}()

	tv := <-wct

	assert.NotEmpty(t, tv)
	assert.Equal(t, 5, tv.Delta)
}

func TestStepWorkerTicker(t *testing.T) {
	t.Run("step increase load duration", func(t *testing.T) {
		wt := StepWorkerTicker{
			C:            make(chan TickValue),
			Start:        5,
			Step:         2,
			Stop:         0,
			StepDuration: 2 * time.Second,
			MaxDuration:  5 * time.Second,
		}

		defer wt.Finish()

		wct := wt.Ticker()

		assert.NotNil(t, wct)

		go func() {
			wt.Run()
		}()

		tv := <-wct
		assert.NotEmpty(t, tv)
		assert.Equal(t, 5, tv.Delta)
		assert.False(t, tv.Done)

		start := time.Now()
		tv = <-wct
		end := time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, 2, tv.Delta)
		assert.False(t, tv.Done)
		expected := 2 * time.Second
		assert.True(t, durationEqual(expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, 2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(2*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		assert.Equal(t, 0, tv.Delta)
		assert.True(t, tv.Done)
	})

	t.Run("step increase load duration with stop", func(t *testing.T) {
		wt := StepWorkerTicker{
			C:            make(chan TickValue),
			Start:        5,
			Step:         2,
			Stop:         15,
			StepDuration: 2 * time.Second,
			MaxDuration:  5 * time.Second,
		}

		defer wt.Finish()

		wct := wt.Ticker()

		assert.NotNil(t, wct)

		go func() {
			wt.Run()
		}()

		tv := <-wct
		assert.NotEmpty(t, tv)
		assert.Equal(t, 5, tv.Delta)
		assert.False(t, tv.Done)

		start := time.Now()
		tv = <-wct
		end := time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, 2, tv.Delta)
		assert.False(t, tv.Done)
		expected := 2 * time.Second
		assert.True(t, durationEqual(expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, 2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(2*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.Equal(t, 6, tv.Delta)
		assert.True(t, tv.Done)
		assert.True(t, durationEqual(3*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)
	})

	t.Run("step decrease load duration", func(t *testing.T) {
		wt := StepWorkerTicker{
			C:            make(chan TickValue),
			Start:        10,
			Step:         -2,
			Stop:         0,
			StepDuration: 2 * time.Second,
			MaxDuration:  5 * time.Second,
		}

		defer wt.Finish()

		wct := wt.Ticker()

		assert.NotNil(t, wct)

		go func() {
			wt.Run()
		}()

		tv := <-wct
		assert.NotEmpty(t, tv)
		assert.Equal(t, 10, tv.Delta)
		assert.False(t, tv.Done)

		start := time.Now()
		tv = <-wct
		end := time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		expected := 2 * time.Second
		assert.True(t, durationEqual(expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(2*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		assert.Equal(t, 0, tv.Delta)
		assert.True(t, tv.Done)
	})

	t.Run("step decrease with stop", func(t *testing.T) {
		wt := StepWorkerTicker{
			C:            make(chan TickValue),
			Start:        10,
			Step:         -2,
			Stop:         4,
			StepDuration: 2 * time.Second,
			MaxDuration:  0,
		}

		defer wt.Finish()

		wct := wt.Ticker()

		assert.NotNil(t, wct)

		go func() {
			wt.Run()
		}()

		tv := <-wct
		assert.NotEmpty(t, tv)
		assert.Equal(t, 10, tv.Delta)
		assert.False(t, tv.Done)

		start := time.Now()
		tv = <-wct
		end := time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		expected := 2 * time.Second
		assert.True(t, durationEqual(expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(2*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(3*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.Equal(t, 0, tv.Delta)
		assert.True(t, tv.Done)
		assert.True(t, durationEqual(4*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)
	})

	t.Run("step decrease with stop and load duration", func(t *testing.T) {

		wt := StepWorkerTicker{
			C:            make(chan TickValue),
			Start:        12,
			Step:         -2,
			Stop:         3,
			StepDuration: 2 * time.Second,
			MaxDuration:  5 * time.Second,
		}

		defer wt.Finish()

		wct := wt.Ticker()

		assert.NotNil(t, wct)

		go func() {
			wt.Run()
		}()

		tv := <-wct
		assert.NotEmpty(t, tv)
		assert.Equal(t, 12, tv.Delta)
		assert.False(t, tv.Done)

		start := time.Now()
		tv = <-wct
		end := time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		expected := 2 * time.Second
		assert.True(t, durationEqual(expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.NotEmpty(t, tv)
		assert.Equal(t, -2, tv.Delta)
		assert.False(t, tv.Done)
		assert.True(t, durationEqual(2*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)

		tv = <-wct
		end = time.Since(start)
		assert.Equal(t, -5, tv.Delta)
		assert.True(t, tv.Done)
		assert.True(t, durationEqual(3*expected, end.Round(time.Second)), "expected %s to equal %s", expected, end)
	})
}
