package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_StatusFromString(t *testing.T) {
	var tests = []struct {
		name     string
		in       string
		expected Status
	}{
		{"OK", "OK", StatusOK},
		{"ok", "ok", StatusOK},
		{"fail", "fail", StatusFail},
		{"FAIL", "FAIL", StatusFail},
		{"asdf", "asdf", StatusOK},
		{"empty", "", StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := StatusFromString(tt.in)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
