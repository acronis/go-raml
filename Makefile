# Directory containing the Makefile.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export PATH := $(GOBIN):$(PATH)

BENCH_FLAGS ?= -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem

# Directories that we want to test and track coverage for.
TEST_DIRS = .

.PHONY: all
all: lint cover

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: test
test:
	@go test ./...

.PHONY: cover
cover:
	@go test -coverprofile=cover.out -coverpkg=./... ./... \
	&& go tool cover -html=cover.out -o cover.html
