BINARY := envy
BUILD_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build test lint fmt-check clean man

build:
	go build -buildvcs=false -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BUILD_DIR)/$(BINARY) ./cmd/envy

test:
	go test -race ./...

lint:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

man:
	@mkdir -p man/
	go run ./cmd/envy doc man/

clean:
	rm -rf $(BUILD_DIR)
