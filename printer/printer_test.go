package printer

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/stretchr/testify/assert"
)

func TestPrinter_getInfluxLine(t *testing.T) {
	date := time.Now()
	unixTimeNow := date.UnixNano()

	var tests = []struct {
		name     string
		report   runner.Report
		expected string
	}{
		{
			"empty",
			runner.Report{},
			`ghz_run,proto="",call="",host="",n=0,c=0,qps=0,z=0,timeout=0,dial_timeout=0,keepalive=0,data="null",metadata="",tags="",errors=0,has_errors=false count=0,total=0,average=0,fastest=0,slowest=0,rps=0.00,errors=0 0`,
		},
		{
			"basic",
			runner.Report{
				Name:      "run name",
				EndReason: runner.ReasonNormalEnd,
				Date:      date,
				Count:     200,
				Total:     time.Duration(2 * time.Second),
				Average:   time.Duration(10 * time.Millisecond),
				Fastest:   time.Duration(1 * time.Millisecond),
				Slowest:   time.Duration(100 * time.Millisecond),
				Rps:       2000,
				ErrorDist: map[string]int{
					"rpc error: code = Internal desc = Internal error.":            3,
					"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2},
				StatusCodeDist: map[string]int{
					"OK":               195,
					"Internal":         3,
					"DeadlineExceeded": 2},
				Options: runner.Options{
					Call:  "helloworld.Greeter.SayHello",
					Proto: "/apis/greeter.proto",
					Host:  "0.0.0.0:50051",
					N:     200,
					C:     50,
					Data: map[string]interface{}{
						"name": "Bob Smith",
					},
					Metadata: &map[string]string{
						"foo bar": "biz baz",
					},
				},
				LatencyDistribution: []runner.LatencyDistribution{
					runner.LatencyDistribution{
						Percentage: 25,
						Latency:    time.Duration(1 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 50,
						Latency:    time.Duration(5 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 75,
						Latency:    time.Duration(10 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 90,
						Latency:    time.Duration(15 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 95,
						Latency:    time.Duration(20 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 99,
						Latency:    time.Duration(25 * time.Millisecond),
					}},
				Histogram: []runner.Bucket{
					runner.Bucket{
						Mark:      0.01,
						Count:     1,
						Frequency: 0.005,
					},
					runner.Bucket{
						Mark:      0.02,
						Count:     10,
						Frequency: 0.01,
					},
					runner.Bucket{
						Mark:      0.03,
						Count:     50,
						Frequency: 0.1,
					},
					runner.Bucket{
						Mark:      0.05,
						Count:     60,
						Frequency: 0.15,
					},
					runner.Bucket{
						Mark:      0.1,
						Count:     15,
						Frequency: 0.07,
					},
				},
				Details: []runner.ResultDetail{
					runner.ResultDetail{
						Timestamp: date,
						Latency:   time.Duration(1 * time.Millisecond),
						Status:    "OK",
					},
				},
			},
			fmt.Sprintf(`ghz_run,name="run\ name",proto="/apis/greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=200,c=50,qps=0,z=0,timeout=0,dial_timeout=0,keepalive=0,data="{\"name\":\"Bob\ Smith\"}",metadata="{\"foo\ bar\":\"biz\ baz\"}",tags="",errors=5,has_errors=true count=200,total=2000000000,average=10000000,fastest=1000000,slowest=100000000,rps=2000.00,median=5000000,p95=20000000,errors=5 %+v`, unixTimeNow),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ReportPrinter{Report: &tt.report}
			actual := p.getInfluxLine()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestPrinter_printInfluxDetails(t *testing.T) {
	date := time.Now()
	unixTimeNow := date.UnixNano()

	var tests = []struct {
		name     string
		report   runner.Report
		expected string
	}{
		{
			"empty",
			runner.Report{},
			``,
		},
		{
			"basic",
			runner.Report{
				Name:      "run name",
				EndReason: runner.ReasonNormalEnd,
				Date:      date,
				Count:     200,
				Total:     time.Duration(2 * time.Second),
				Average:   time.Duration(10 * time.Millisecond),
				Fastest:   time.Duration(1 * time.Millisecond),
				Slowest:   time.Duration(100 * time.Millisecond),
				Rps:       2000,
				ErrorDist: map[string]int{
					"rpc error: code = Internal desc = Internal error.":            3,
					"rpc error: code = DeadlineExceeded desc = Deadline exceeded.": 2},
				StatusCodeDist: map[string]int{
					"OK":               195,
					"Internal":         3,
					"DeadlineExceeded": 2},
				Options: runner.Options{
					Call:  "helloworld.Greeter.SayHello",
					Proto: "/apis/greeter.proto",
					Host:  "0.0.0.0:50051",
					N:     200,
					C:     50,
					Data: map[string]interface{}{
						"name": "Bob Smith",
					},
					Metadata: &map[string]string{
						"foo bar": "biz baz",
					},
				},
				LatencyDistribution: []runner.LatencyDistribution{
					runner.LatencyDistribution{
						Percentage: 25,
						Latency:    time.Duration(1 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 50,
						Latency:    time.Duration(5 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 75,
						Latency:    time.Duration(10 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 90,
						Latency:    time.Duration(15 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 95,
						Latency:    time.Duration(20 * time.Millisecond),
					},
					runner.LatencyDistribution{
						Percentage: 99,
						Latency:    time.Duration(25 * time.Millisecond),
					}},
				Histogram: []runner.Bucket{
					runner.Bucket{
						Mark:      0.01,
						Count:     1,
						Frequency: 0.005,
					},
					runner.Bucket{
						Mark:      0.02,
						Count:     10,
						Frequency: 0.01,
					},
					runner.Bucket{
						Mark:      0.03,
						Count:     50,
						Frequency: 0.1,
					},
					runner.Bucket{
						Mark:      0.05,
						Count:     60,
						Frequency: 0.15,
					},
					runner.Bucket{
						Mark:      0.1,
						Count:     15,
						Frequency: 0.07,
					},
				},
				Details: []runner.ResultDetail{
					runner.ResultDetail{
						Timestamp: date,
						Latency:   time.Duration(1 * time.Millisecond),
						Status:    "OK",
					},
				},
			},
			fmt.Sprintf(`ghz_detail,name="run\ name",proto="/apis/greeter.proto",call="helloworld.Greeter.SayHello",host="0.0.0.0:50051",n=200,c=50,qps=0,z=0,timeout=0,dial_timeout=0,keepalive=0,data="{\"name\":\"Bob\ Smith\"}",metadata="{\"foo\ bar\":\"biz\ baz\"}",tags="",hasError=false latency=1000000,error="",status="OK" %+v
`, unixTimeNow),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString("")
			p := ReportPrinter{Report: &tt.report, Out: buf}
			p.printInfluxDetails()
			actual := buf.String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
