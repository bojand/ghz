package printer

import (
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
				Name:           "run name",
				EndReason:      runner.ReasonNormalEnd,
				Date:           date,
				Count:          200,
				Total:          time.Duration(100 * time.Millisecond),
				ErrorDist:      make(map[string]int),
				StatusCodeDist: make(map[string]int),
				Options:        &runner.Options{}},
			fmt.Sprintf(`ghz_run,name="run\ name",proto="",call="",host="",n=0,c=0,qps=0,z=0,timeout=0,dial_timeout=0,keepalive=0,data="null",metadata="",tags="",errors=0,has_errors=false count=200,total=100000000,average=0,fastest=0,slowest=0,rps=0.00,errors=0 %+v`, unixTimeNow),
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
