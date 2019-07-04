// +build tools

// Package tools blank imports Golang tools we use as part of the build.
package tools

import (
	_ "http://golang.org/x/lint/golint"      // tool
	_ "github.com/kisielk/errcheck"        // tool
	_ "honnef.co/go/tools/cmd/staticcheck" // tool
)
