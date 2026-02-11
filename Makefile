BINARY := envy
BUILD_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test lint fmt-check clean

build:
	go build -buildvcs=false -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) ./cmd/envy

test:
	go test -race ./...

lint:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

clean:
	rm -rf $(BUILD_DIR)
