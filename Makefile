BINARY := envy
BUILD_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test lint clean

build:
	go build -buildvcs=false -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY) ./cmd/envy

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
