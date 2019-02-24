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

# Set OPEN_COVERAGE=1 to open the coverage.html file after running make cover.
ifeq ($(OPEN_COVERAGE),1)
	OPEN_COVERAGE_HTML := 1
else
	OPEN_COVERAGE_HTML :=
endif

#################################################
##### Everything below should not be edited #####
#################################################

# UNAME_OS stores the value of uname -s.
UNAME_OS := $(shell uname -s)
# UNAME_ARCH stores the value of uname -m.
UNAME_ARCH := $(shell uname -m)

# TMP_BASE is the base directory used for TMP.
# Use TMP and not TMP_BASE as the temporary directory.
TMP_BASE := .tmp
# TMP is the temporary directory used.
# This is based on UNAME_OS and UNAME_ARCH to make sure there are no issues
# switching between platform builds.
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
# TMP_BIN is where we install binaries for usage during development.
TMP_BIN = $(TMP)/bin
# TMP_COVERAGE is where we store code coverage files.
TMP_COVERAGE := $(TMP_BASE)/coverage

# The following unexports and exports put us into Golang Modules mode.

# Make sure GOPATH is unset so that it is not possible to interfere with other packages.
unexport GOPATH
# Make sure GOROOT is unset so that there are no issues.
unexport GOROOT
# Turn Golang modules on
export GO111MODULE := on
# Set the location where we install Golang binaries via go install to TMP_BIN.
export GOBIN := $(abspath $(TMP_BIN))
# Add GOBIN to to the front of the PATH. This allows us to invoke binaries we install.
export PATH := :$(GOBIN):$(PATH)

# GO_MODULE extracts the module name from the go.mod file.
GO_MODULE := $(shell grep '^module ' go.mod | cut -f 2 -d ' ')

# Run all by default when "make" is invoked.
.DEFAULT_GOAL := all

# All runs the default lint, test, and code coverage targets.
.PHONY: all
all: lint cover

# Clean removes all temporary files.
.PHONY: clean
clean:
	rm -rf $(TMP_BASE)

# Golint runs the golint linter.
.PHONY: golint
golint:
	$(AT) go install github.com/golang/lint/golint
	golint -set_exit_status $(GO_PKGS)

# Errcheck runs the errcheck linter.
.PHONY: errcheck
errcheck:
	$(AT) go install github.com/kisielk/errcheck
	errcheck -ignoretests $(GO_PKGS)

# Staticcheck runs the staticcheck linter.
.PHONY: staticcheck
staticcheck:
	$(AT) go install honnef.co/go/tools/cmd/staticcheck
	staticcheck --tests=false $(GO_PKGS)

# Lint runs all linters. This is the main lint target to run.
# TODO: add errcheck and staticcheck when the code is updated to pass them
.PHONY: lint
lint: golint errcheck

# Test runs go test on GO_PKGS. This does not produce code coverage.
.PHONY: test
test:
	go test $(GO_TEST_FLAGS) $(GO_PKGS)

# Cover runs go_test on GO_PKGS and produces code coverage in multiple formats.
# A coverage.html file for human viewing will be at $(TMP_COVERAGE)/coverage.html
# This target will echo "open $(TMP_COVERAGE)/coverage.html" with TMP_COVERAGE
# expanded so that you can easily copy "open $(TMP_COVERAGE)/coverage.html" into
# your terminal as a command to run, and then see the code coverage output locally.
.PHONY: cover
cover:
	$(AT) rm -rf $(TMP_COVERAGE)
	$(AT) mkdir -p $(TMP_COVERAGE)
	go test $(GO_TEST_FLAGS) -coverprofile=$(TMP_COVERAGE)/coverage.txt -coverpkg=$(shell echo $(GO_PKGS) | grep -v \/cmd\/ | tr ' ' ',') $(GO_PKGS)
	$(AT) go tool cover -html=$(TMP_COVERAGE)/coverage.txt -o $(TMP_COVERAGE)/coverage.html
	$(AT) echo
	$(AT) go tool cover -func=$(TMP_COVERAGE)/coverage.txt | grep total
	$(AT) echo
	$(AT) echo Open the coverage report:
	$(AT) echo open $(TMP_COVERAGE)/coverage.html
	$(AT) if [ "$(OPEN_COVERAGE_HTML)" == "1" ]; then open $(TMP_COVERAGE)/coverage.html; fi
