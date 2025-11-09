VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build run test clean version lint fmt

build:
	go build -ldflags "$(LDFLAGS)" ./...

run:
	go run -ldflags "$(LDFLAGS)" .

test:
	go test ./...

clean:
	rm -f wtm

version:
	@echo $(VERSION)

lint:
	golangci-lint run

fmt:
	golangci-lint run --fix
