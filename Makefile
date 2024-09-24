# Directory containing the Makefile.
export PATH := $(GOBIN):$(PATH)

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

.PHONY: build
build: go-build

.PHONY: go-build
go-build:
	@cd cmd/raml && go build -o ../../.build/raml

.PHONY: install
install: go-install

.PHONY: go-install
go-install:
	@cd cmd/raml && \
	go install -v ./... \
	&& echo `go list -f '{{.Module.Path}}'` has been installed to `go list -f '{{.Target}}'` && true
