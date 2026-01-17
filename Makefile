VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)
GOBIN ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: build run test clean version lint fmt install install-git

build:
	go build -ldflags "$(LDFLAGS)" -o wtm .

run:
	go run -ldflags "$(LDFLAGS)" .

test:
	go test ./...

install:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/wtm .

install-git:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/git-wtm .

clean:
	rm -f wtm git-wtm

version:
	@echo $(VERSION)

lint:
	golangci-lint run

fmt:
	golangci-lint run --fix
