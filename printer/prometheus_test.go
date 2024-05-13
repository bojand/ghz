package printer

import (
	"bytes"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	promtypes "github.com/prometheus/client_model/go"
	expfmt "github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
)

func TestPrinter_printPrometheus(t *testing.T) {
	date := time.Now()

	var tests = []struct {
		name     string
		report   runner.Report
		expected string
	}{
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
					Name:         "run name",
					Call:         "helloworld.Greeter.SayHello",
					Proto:        "/apis/greeter.proto",
					Host:         "0.0.0.0:50051",
					LoadSchedule: "const",
					CSchedule:    "const",
					Total:        200,
					Concurrency:  50,
					Data: map[string]interface{}{
						"name": "Bob Smith",
					},
					Metadata: &map[string]string{
						"foo bar": "biz baz",
					},
				},
				LatencyDistribution: []runner.LatencyDistribution{
					{
						Percentage: 25,
						Latency:    time.Duration(1 * time.Millisecond),
					},
					{
						Percentage: 50,
						Latency:    time.Duration(5 * time.Millisecond),
					},
					{
						Percentage: 75,
						Latency:    time.Duration(10 * time.Millisecond),
					},
					{
						Percentage: 90,
						Latency:    time.Duration(15 * time.Millisecond),
					},
					{
						Percentage: 95,
						Latency:    time.Duration(20 * time.Millisecond),
					},
					{
						Percentage: 99,
						Latency:    time.Duration(25 * time.Millisecond),
					}},
				Histogram: []runner.Bucket{
					{
						Mark:      0.01,
						Count:     1,
						Frequency: 0.005,
					},
					{
						Mark:      0.02,
						Count:     10,
						Frequency: 0.01,
					},
					{
						Mark:      0.03,
						Count:     50,
						Frequency: 0.1,
					},
					{
						Mark:      0.05,
						Count:     60,
						Frequency: 0.15,
					},
					{
						Mark:      0.1,
						Count:     15,
						Frequency: 0.07,
					},
				},
				Details: []runner.ResultDetail{
					{
						Timestamp: date,
						Latency:   time.Duration(1 * time.Millisecond),
						Status:    "OK",
					},
				},
			},
			expectedProm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString("")
			p := ReportPrinter{Report: &tt.report, Out: buf}
			err := p.printPrometheus()
			assert.NoError(t, err)
			actual := buf.String()

			// parse actual
			var actualMetricFamilies []*promtypes.MetricFamily
			r := bytes.NewReader([]byte(actual))
			decoder := expfmt.NewDecoder(r, expfmt.NewFormat(expfmt.TypeTextPlain))
			for {
				metric := &promtypes.MetricFamily{}
				err := decoder.Decode(metric)
				if err != nil {
					if err == io.EOF {
						break
					}
					assert.NoError(t, err)
				}

				actualMetricFamilies = append(actualMetricFamilies, metric)
			}

			// parse expected
			var expectedMetricFamilies []*promtypes.MetricFamily
			r = bytes.NewReader([]byte(tt.expected))
			decoder = expfmt.NewDecoder(r, expfmt.NewFormat(expfmt.TypeTextPlain))
			for {
				metric := &promtypes.MetricFamily{}
				err := decoder.Decode(metric)
				if err != nil {
					if err == io.EOF {
						break
					}
					assert.NoError(t, err)
				}

				expectedMetricFamilies = append(expectedMetricFamilies, metric)
			}

			for i, amf := range actualMetricFamilies {
				amf := amf

				assert.True(t, i < len(expectedMetricFamilies))

				emf := expectedMetricFamilies[i]

				for im, am := range amf.Metric {
					am := am

					assert.True(t, im < len(emf.Metric))

					em := emf.Metric[im]

					assert.NotNil(t, em)

					// sort actual labels
					al := am.Label
					sort.Slice(al, func(i, j int) bool { return *(al[i].Name) < *(al[j].Name) })

					// sort expected labels
					el := em.Label
					sort.Slice(el, func(i, j int) bool { return *(el[i].Name) < *(el[j].Name) })

					// finally compare labels
					assert.Equal(t, el, al)
				}
			}
		})
	}
}

var expectedProm string = `
# TYPE ghz_run_count gauge
ghz_run_count{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 200
# TYPE ghz_run_total gauge
ghz_run_total{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 2e+09
# TYPE ghz_run_average gauge
ghz_run_average{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 1e+07
# TYPE ghz_run_fastest gauge
ghz_run_fastest{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 1e+06
# TYPE ghz_run_slowest gauge
ghz_run_slowest{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 1e+08
# TYPE ghz_run_rps gauge
ghz_run_rps{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 2000
# TYPE ghz_run_histogram histogram
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="0.01"} 1
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="0.02"} 10
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="0.03"} 50
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="0.05"} 60
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="0.1"} 15
ghz_run_histogram_bucket{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",le="+Inf"} 200
ghz_run_histogram_sum{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 2e+09
ghz_run_histogram_count{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 200
# TYPE ghz_run_latency summary
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.25"} 1e+06
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.5"} 5e+06
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.75"} 1e+07
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.9"} 1.5e+07
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.95"} 2e+07
ghz_run_latency{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0",quantile="0.99"} 2.5e+07
ghz_run_latency_sum{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 2e+09
ghz_run_latency_count{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 200
# TYPE ghz_run_errors gauge
ghz_run_errors{name="run name",end_reason="normal",insecure="false",rps="0",connections="0",keepalive="0",skipFirst="0",dial_timeout="0",proto="/apis/greeter.proto",concurrency="50",call="helloworld.Greeter.SayHello",import_paths="",async="false",binary="false",total="200",host="0.0.0.0:50051",skipTLS="false",CPUs="0",timeout="0",count_errors="false",duration="0"} 5
`
