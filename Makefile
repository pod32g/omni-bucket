VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/pod32g/omni-bucket/internal/version.Version=$(VERSION)

.PHONY: build test vet

build:
	go build -ldflags "$(LDFLAGS)" -o bb ./cmd/omni-bucket

test:
	go test ./...

vet:
	go vet ./...
