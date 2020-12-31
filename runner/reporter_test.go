package runner

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReport_MarshalJSON(t *testing.T) {
	z, _ := time.Parse(time.RFC822Z, "02 Jan 06 15:04 -0700")
	r := &Report{
		Date:    z,
		Count:   1000,
		Total:   time.Duration(10) * time.Second,
		Average: time.Duration(500) * time.Millisecond,
		Fastest: time.Duration(10) * time.Millisecond,
		Slowest: time.Duration(1000) * time.Millisecond,
		Rps:     34567.89,
	}

	json, err := json.Marshal(&r)
	assert.NoError(t, err)

	expected := `{"date":"2006-01-02T15:04:00-07:00","options":{"insecure":false,"load-schedule":"","load-start":0,"load-end":0,"load-step":0,"load-step-duration":0,"load-max-duration":0,"concurrency-schedule":"","concurrency-start":0,"concurrency-end":0,"concurrency-step":0,"concurrency-step-duration":0,"concurrency-max-duration":0,"binary":false,"CPUs":0},"count":1000,"total":10000000000,"average":500000000,"fastest":10000000,"slowest":1000000000,"rps":34567.89,"errorDistribution":null,"statusCodeDistribution":null,"latencyDistribution":null,"histogram":null,"details":null}`
	assert.Equal(t, expected, string(json))
}

func TestReport_CorrectDetails(t *testing.T) {
	callResultsChan := make(chan *callResult)
	config, _ := NewConfig("call", "host")
	reporter := newReporter(callResultsChan, config)

	go reporter.Run()

	cr1 := callResult{
		status:    "OK",
		duration:  time.Millisecond * 100,
		err:       nil,
		timestamp: time.Now(),
	}
	callResultsChan <- &cr1
	cr2 := callResult{
		status:    "DeadlineExceeded",
		duration:  time.Millisecond * 500,
		err:       context.DeadlineExceeded,
		timestamp: time.Now(),
	}
	callResultsChan <- &cr2

	close(callResultsChan)
	<-reporter.done
	report := reporter.Finalize("stop reason", time.Second)

	assert.Equal(t, 2, len(report.Details))
	assert.Equal(t, ResultDetail{Error: "", Latency: cr1.duration, Status: cr1.status, Timestamp: cr1.timestamp}, report.Details[0])
	assert.Equal(t, ResultDetail{Error: cr2.err.Error(), Latency: cr2.duration, Status: cr2.status, Timestamp: cr2.timestamp}, report.Details[1])
}

func TestReport_latencies(t *testing.T) {
	var tests = []struct {
		input    []float64
		expected []LatencyDistribution
	}{
		{
			input: []float64{15, 20, 35, 40, 50},
			expected: []LatencyDistribution{
				{Percentage: 10, Latency: 15 * time.Second},
				{Percentage: 25, Latency: 20 * time.Second},
				{Percentage: 50, Latency: 35 * time.Second},
				{Percentage: 75, Latency: 40 * time.Second},
				{Percentage: 90, Latency: 50 * time.Second},
				{Percentage: 95, Latency: 50 * time.Second},
				{Percentage: 99, Latency: 50 * time.Second},
			},
		},
		{
			input: []float64{3, 6, 7, 8, 8, 10, 13, 15, 16, 20},
			expected: []LatencyDistribution{
				{Percentage: 10, Latency: 3 * time.Second},
				{Percentage: 25, Latency: 7 * time.Second},
				{Percentage: 50, Latency: 8 * time.Second},
				{Percentage: 75, Latency: 15 * time.Second},
				{Percentage: 90, Latency: 16 * time.Second},
				{Percentage: 95, Latency: 20 * time.Second},
				{Percentage: 99, Latency: 20 * time.Second},
			},
		},
		{
			input: []float64{3, 6, 7, 8, 8, 9, 10, 13, 15, 16, 20},
			expected: []LatencyDistribution{
				{Percentage: 10, Latency: 6 * time.Second},
				{Percentage: 25, Latency: 7 * time.Second},
				{Percentage: 50, Latency: 9 * time.Second},
				{Percentage: 75, Latency: 15 * time.Second},
				{Percentage: 90, Latency: 16 * time.Second},
				{Percentage: 95, Latency: 20 * time.Second},
				{Percentage: 99, Latency: 20 * time.Second},
			},
		},
		{
			input: []float64{2.1, 3.2, 4.5, 6.3, 7.4, 8.5, 9.6, 10.7, 13.8, 15.9, 16.11, 18.17, 20.11, 22.34},
			expected: []LatencyDistribution{
				{Percentage: 10, Latency: time.Duration(3.2 * float64(time.Second))},
				{Percentage: 25, Latency: time.Duration(6.3 * float64(time.Second))},
				{Percentage: 50, Latency: time.Duration(9.6 * float64(time.Second))},
				{Percentage: 75, Latency: time.Duration(16.11 * float64(time.Second))},
				{Percentage: 90, Latency: time.Duration(20.11 * float64(time.Second))},
				{Percentage: 95, Latency: time.Duration(22.34 * float64(time.Second))},
				{Percentage: 99, Latency: time.Duration(22.34 * float64(time.Second))},
			},
		},
	}

	for i, tt := range tests {
		t.Run("latencies "+strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			lats := latencies(tt.input)
			assert.Equal(t, tt.expected, lats)
		})
	}
}
