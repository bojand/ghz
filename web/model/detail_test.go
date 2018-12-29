package model

import (
	"testing"
	"time"

	"github.com/bojand/ghz/runner"
	"github.com/stretchr/testify/assert"
)

func TestDetail_BeforeSave(t *testing.T) {
	var details = []struct {
		name        string
		in          *Detail
		expected    *Detail
		expectError bool
	}{
		{"no run", &Detail{ResultDetail: runner.ResultDetail{Latency: 12345}}, &Detail{ResultDetail: runner.ResultDetail{Latency: 12345}}, true},
		{"trim error", &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: " network error "}}, &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: "network error", Status: "OK"}}, false},
		{"trim status", &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: " network error ", Status: " OK "}}, &Detail{ReportID: 1, ResultDetail: runner.ResultDetail{Latency: 12345, Error: "network error", Status: "OK"}}, false},
	}

	for _, tt := range details {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.BeforeSave(nil)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, tt.in)
		})
	}
}

func TestDetail_UnmarshalJSON(t *testing.T) {
	expectedTime, err := time.Parse("2006-01-02T15:04:05-0700", "2018-08-08T13:00:00-0300")
	assert.NoError(t, err)

	var details = []struct {
		name        string
		in          string
		expected    *Detail
		expectError bool
	}{
		{"RFC3339",
			`{"timestamp":"2018-08-08T13:00:00.000000000-03:00","latency":123,"error":"","status":"OK"}`,
			&Detail{ResultDetail: runner.ResultDetail{Timestamp: expectedTime, Latency: 123, Error: "", Status: "OK"}},
			false},
		{"layoutISO2",
			`{"timestamp":"2018-08-08T13:00:00-0300","latency":123,"error":"","status":"OK"}`,
			&Detail{ResultDetail: runner.ResultDetail{Timestamp: expectedTime, Latency: 123, Error: "", Status: "OK"}},
			false},
	}

	for _, tt := range details {
		t.Run(tt.name, func(t *testing.T) {
			var d Detail
			err := d.UnmarshalJSON([]byte(tt.in))
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, &d)
		})
	}
}
