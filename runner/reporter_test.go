package runner

import (
	"context"
	"encoding/json"
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

	expected := `{"date":"2006-01-02T15:04:00-07:00","options":{"insecure":false,"binary":false,"CPUs":0},"count":1000,"total":10000000000,"average":500000000,"fastest":10000000,"slowest":1000000000,"rps":34567.89,"errorDistribution":null,"statusCodeDistribution":null,"latencyDistribution":null,"histogram":null,"details":null}`
	assert.Equal(t, expected, string(json))
}

func TestReport_CorrectDetails(t *testing.T) {
	callResultsChan := make(chan *callResult)
	config, _ := newConfig("call", "host")
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
