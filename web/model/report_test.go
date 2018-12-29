package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport_BeforeSave(t *testing.T) {
	var reports = []struct {
		name        string
		in          *Report
		expected    *Report
		expectError bool
	}{
		{"no project id", &Report{}, &Report{}, true},
		{"with project id", &Report{ProjectID: 123}, &Report{ProjectID: 123, Status: "ok"}, false},
	}

	for _, tt := range reports {
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
