package ghz

import (
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

	expected := `{"date":"2006-01-02T15:04:00-07:00","count":1000,"total":10000000000,"average":500000000,"fastest":10000000,"slowest":1000000000,"rps":34567.89,"errorDistribution":null,"statusCodeDistribution":null,"latencyDistribution":null,"histogram":null,"details":null}`
	assert.Equal(t, expected, string(json))
}
