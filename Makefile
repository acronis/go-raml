# Directory containing the Makefile.
export PATH := $(GOBIN):$(PATH)

GO_VERSION = 1.20.14
GO_VERSION_CMD_RAML = 1.22.10
GOLANGCI_LINT_VERSION = 1.55.2

.PHONY: all
all: lint cover

.PHONY: lint
lint: go-install
	@go$(GO_VERSION) run github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION) run ./...

.PHONY: test
test-unit: go-install
	@go$(GO_VERSION) test ./...

# the coverage should be at least 80% for now, but it should be increased in the future to 90% and more
.PHONY: cover
test-cover: go-install
	@pkgs=$$(go$(GO_VERSION) list ./... | grep -v /Store/) \
	go$(GO_VERSION) test -coverprofile=cover.out.tmp $${pkgs} -coverpkg=$${pkgs}  \
	&& cat cover.out.tmp | grep -v "rdtparser_base_visitor.go" > cover.out \
	&& rm cover.out.tmp \
	&& go$(GO_VERSION) tool cover -func=cover.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' | \
	awk '{if ($$1 < 80) {print "Coverage is below 80%!" ; exit 1}}' \
	&& go$(GO_VERSION) tool cover -html=cover.out -o cover.html

.PHONY: test
test: test-cover

.PHONY: build
build: go-build

.PHONY: go-build
go-build: go-install-cmd-raml
	@echo "Building raml"
	@cd cmd/raml && go$(GO_VERSION_CMD_RAML) build -o ../../.build/raml \
	&& echo "Build successful in .build/raml using go$(GO_VERSION_CMD_RAML)"

.PHONY: go-install
go-install:
	@if ! which go$(GO_VERSION) > /dev/null 2>&1; then \
	echo "Installing go$(GO_VERSION)"; \
	go install golang.org/dl/go$(GO_VERSION)@latest; \
	go$(GO_VERSION) download; \
	fi

.PHONY: go-install-cmd-raml
go-install-cmd-raml:
	@$(MAKE) go-install GO_VERSION=$(GO_VERSION_CMD_RAML)

.PHONY: go
go: go-install
	@# Run go command with the specified arguments
	@# examples:
	@# 		make go a="mod tidy"
	@# 		make go a="mod vendor"
	@go$(GO_VERSION) $(a)

.PHONY: install
install: go-install-cmd-raml
	@echo "Installing raml"
	@cd cmd/raml && go$(GO_VERSION_CMD_RAML) install . \
	&& echo "Installed using go$(GO_VERSION_CMD_RAML) to $(GOPATH)/bin/raml"