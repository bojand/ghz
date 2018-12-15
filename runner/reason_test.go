package runner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReason_String(t *testing.T) {
	var tests = []struct {
		name     string
		in       StopReason
		expected string
	}{
		{"normal", ReasonNormalEnd, "normal"},
		{"cancel", ReasonCancel, "cancel"},
		{"timeout", ReasonTimeout, "timeout"},
		{"unknown", StopReason("foo"), "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.in.String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestReason_StatusFromString(t *testing.T) {
	var tests = []struct {
		name     string
		in       string
		expected StopReason
	}{
		{"normal", "normal", ReasonNormalEnd},
		{"cancel", "cancel", ReasonCancel},
		{"timeout", "timeout", ReasonTimeout},
		{"unknown", "foo", ReasonNormalEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ReasonFromString(tt.in)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestReason_UnmarshalJSON(t *testing.T) {
	var tests = []struct {
		name     string
		in       string
		expected StopReason
	}{
		{"normal", `"normal"`, ReasonNormalEnd},
		{"NORMAL", `"NORMAL"`, ReasonNormalEnd},
		{"cancel", `"cancel"`, ReasonCancel},
		{"CANCEL", `"CANCEL"`, ReasonCancel},
		{" CANCEL ", ` "CANCEL" `, ReasonCancel},
		{"timeout", ` "timeout" `, ReasonTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual StopReason
			err := json.Unmarshal([]byte(tt.in), &actual)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestReason_MarshalJSON(t *testing.T) {
	var tests = []struct {
		name     string
		in       StopReason
		expected string
	}{
		{"normal", ReasonNormalEnd, `"normal"`},
		{"cancel", ReasonCancel, `"cancel"`},
		{"timeout", ReasonTimeout, `"timeout"`},
		{"unknown", StopReason("foo"), `"normal"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := json.Marshal(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(actual))
		})
	}
}
