VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: build run test clean version lint fmt install install-git release-prep

build:
	go build -ldflags "$(LDFLAGS)" -o dist/wtm ./cmd/wtm

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/wtm

test:
	go test ./...

install:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/wtm ./cmd/wtm

install-git:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/git-wtm ./cmd/wtm

clean:
	rm -rf dist

version:
	@echo $(VERSION)

lint:
	golangci-lint run

fmt:
	golangci-lint run --fix

release-prep:
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is required. Usage: make release-prep RELEASE_VERSION=0.10.0)
endif
	@perl -pi -e 's/"version":\s*"[^"]*"/"version": "$(RELEASE_VERSION)"/' .claude-plugin/marketplace.json
	@echo "Updated .claude-plugin/marketplace.json to $(RELEASE_VERSION)"
