BINARY := envy
BUILD_DIR := bin

.PHONY: build test lint clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/envy

test:
	go test -race ./...

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
