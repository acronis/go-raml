# Directory containing the Makefile.
export PATH := $(GOBIN):$(PATH)

.PHONY: all
all: lint cover

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: test
test-unit:
	@go test ./...

# the coverage should be at least 78% for now, but it should be increased in the future to 90% and more
.PHONY: cover
test-cover:
	@go test -coverprofile=cover.out -coverpkg=./... ./... \
	&& go tool cover -func=cover.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' | \
	awk '{if ($$1 < 78) {print "Coverage is below 78%!" ; exit 1}}' \
	&& go tool cover -html=cover.out -o cover.html

.PHONY: test
test: test-cover

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
	&& echo $$(go list -f '{{.Module.Path}}') has been installed to $$(go list -f '{{.Target}}') && true