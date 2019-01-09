package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_BeforeSave(t *testing.T) {
	var hs = []struct {
		name        string
		in          *Options
		expected    *Options
		expectError bool
	}{
		{"no report id", &Options{}, &Options{}, true},
		{"with report id", &Options{ReportID: 123}, &Options{ReportID: 123}, false},
	}

	for _, tt := range hs {
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
