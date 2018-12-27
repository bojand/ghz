package model

import "strings"

// Status represents a status of a project or record
type Status string

// StatusFromString creates a Status from a string
func StatusFromString(str string) Status {
	str = strings.ToLower(str)

	t := StatusOK

	if str == "fail" {
		t = StatusFail
	}

	return t
}

const (
	// StatusOK means the latest run in test was within the threshold
	StatusOK = Status("ok")

	// StatusFail means the latest run in test was not within the threshold
	StatusFail = Status("fail")
)
