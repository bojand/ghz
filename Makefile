# SHELL defines the shell that the Makefile uses.
# We also set -o pipefail so that if a previous command in a pipeline fails, a command fails.
# http://redsymbol.net/articles/unofficial-bash-strict-mode
SHELL := /bin/bash -o pipefail

############### RUNTIME VARIABLES ###############
# The following variables can be set by the user
# To adjust the functionality of the targets.
#################################################

# Set V=1 on the command line to turn off all suppression. Many trivial
# commands are suppressed with "@", by setting V=1, this will be turned off.
ifeq ($(V),1)
	AT :=
else
	AT := @
endif

# GO_PKGS is the list of packages to run our linting and testing commands against.
# This can be set when invoking a target.
GO_PKGS ?= $(shell go list ./...)

# GO_TEST_FLAGS are the flags passed to go test
GO_TEST_FLAGS ?= -race

# directory to output build
DIST_DIR=./dist

# export GOPATH
export GOPATH := $(shell go env GOPATH)

# Set OPEN_COVERAGE=1 to open the coverage.html file after running make cover.
ifeq ($(OPEN_COVERAGE),1)
	OPEN_COVERAGE_HTML := 1
else
	OPEN_COVERAGE_HTML :=
endif

# UNAME_OS stores the value of uname -s.
UNAME_OS := $(shell uname -s)
# UNAME_ARCH stores the value of uname -m.
UNAME_ARCH := $(shell uname -m)

# TMP_BASE is the base directory used for TMP.
# Use TMP and not TMP_BASE as the temporary directory.
TMP_BASE := .tmp
# TMP_COVERAGE is where we store code coverage files.
TMP_COVERAGE := $(TMP_BASE)/coverage

# Run all by default when "make" is invoked.
.DEFAULT_GOAL := all

# Tools
.PHONY: tools
tools:
	go generate -tags tools tools/tools.go

# All runs the default lint, test, and code coverage targets.
.PHONY: all
all: lint cover build

# Clean removes all temporary files.
.PHONY: clean
clean:
	rm -rf $(TMP_BASE)
	rm -rf dist

# Lint runs all linters. This is the main lint target to run.
.PHONY: lint
lint: 
	golangci-lint run ./...

# Test runs go test on GO_PKGS. This does not produce code coverage.
.PHONY: test
test:
	go test $(GO_TEST_FLAGS) $(GO_PKGS)

# Formats using gofmt and goimports all go files
.PHONY: fmt
fmt:
	find . -name '*.go' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

# Build
.PHONY: build
build:
	CGO_ENABLED=0 go build --ldflags="-s -w" -o $(DIST_DIR)/ghz ./cmd/ghz/...
	go build --ldflags="-s -w" -o $(DIST_DIR)/ghz-web ./cmd/ghz-web/...

# Cover runs go_test on GO_PKGS and produces code coverage in multiple formats.
# A coverage.html file for human viewing will be at $(TMP_COVERAGE)/coverage.html
# This target will echo "open $(TMP_COVERAGE)/coverage.html" with TMP_COVERAGE
# expanded so that you can easily copy "open $(TMP_COVERAGE)/coverage.html" into
# your terminal as a command to run, and then see the code coverage output locally.
.PHONY: cover
cover:
	$(AT) rm -rf $(TMP_COVERAGE)
	$(AT) mkdir -p $(TMP_COVERAGE)
	go test $(GO_TEST_FLAGS) -json -cover -coverprofile=$(TMP_COVERAGE)/coverage.txt $(GO_PKGS) | tparse
	$(AT) go tool cover -html=$(TMP_COVERAGE)/coverage.txt -o $(TMP_COVERAGE)/coverage.html
	$(AT) echo
	$(AT) go tool cover -func=$(TMP_COVERAGE)/coverage.txt | grep total
	$(AT) echo
	$(AT) echo Open the coverage report:
	$(AT) echo open $(TMP_COVERAGE)/coverage.html
	$(AT) if [ "$(OPEN_COVERAGE_HTML)" == "1" ]; then open $(TMP_COVERAGE)/coverage.html; fi
