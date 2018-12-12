package runner

import (
	"fmt"
	"strings"
)

// StopReason is a reason why the run ended
type StopReason string

// String() is the string representation of threshold
func (s StopReason) String() string {
	if s == ReasonCancel {
		return "cancel"
	}

	if s == ReasonTimeout {
		return "timeout"
	}

	return "normal"
}

// UnmarshalJSON prases a Threshold value from JSON string
func (s *StopReason) UnmarshalJSON(b []byte) error {
	input := strings.TrimLeft(string(b), `"`)
	input = strings.TrimRight(input, `"`)
	*s = ReasonFromString(input)
	return nil
}

// MarshalJSON formats a Threshold value into a JSON string
func (s StopReason) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", s.String())), nil
}

// ReasonFromString creates a Status from a string
func ReasonFromString(str string) StopReason {
	str = strings.ToLower(str)

	s := ReasonNormalEnd

	if str == "cancel" {
		s = ReasonCancel
	}

	if str == "timeout" {
		s = ReasonTimeout
	}

	return s
}

const (
	// ReasonNormalEnd indicates a normal end to the run
	ReasonNormalEnd = StopReason("normal")

	// ReasonCancel indicates end due to cancellation
	ReasonCancel = StopReason("cancel")

	// ReasonTimeout indicates run ended due to Z parameter timeout
	ReasonTimeout = StopReason("timeout")
)
