# Directory containing the Makefile.
export PATH := $(GOBIN):$(PATH)

GO_VERISON = 1.20.14
GOLANGCI_LINT_VERSION = 1.55.2

.PHONY: all
all: lint cover

.PHONY: lint
lint:
	@go$(GO_VERISON) run github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) run ./...

.PHONY: test
test-unit:
	@go test ./...

# the coverage should be at least 80% for now, but it should be increased in the future to 90% and more
.PHONY: cover
test-cover:
	@pkgs=$$(go list ./... | grep -v /Store/) \
	go test -coverprofile=cover.out.tmp $${pkgs} -coverpkg=$${pkgs}  \
	&& cat cover.out.tmp | grep -v "rdtparser_base_visitor.go" > cover.out \
	&& rm cover.out.tmp \
	&& go tool cover -func=cover.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' | \
	awk '{if ($$1 < 80) {print "Coverage is below 80%!" ; exit 1}}' \
	&& go tool cover -html=cover.out -o cover.html

.PHONY: test
test: test-cover

.PHONY: build
build: go-build

.PHONY: go-build
go-build:
	@cd cmd/raml && go build -o ../../.build/raml


.PHONY: go-install
go-install:
	go install golang.org/dl/go$(GO_VERISON)@latest
	go$(GO_VERISON) download
