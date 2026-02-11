BINARY := envy
BUILD_DIR := bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.DEFAULT_GOAL := help
.PHONY: help build test lint fmt-check clean man release

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Build binary to bin/envy
	go build -buildvcs=false -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BUILD_DIR)/$(BINARY) ./cmd/envy

test: ## Run all tests with race detector
	go test -race ./...

lint: ## Run go vet
	go vet ./...

fmt-check: ## Check code formatting
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:"; gofmt -l .; exit 1)

man: ## Generate man pages
	@mkdir -p man/
	go run ./cmd/envy doc man/

clean: ## Remove build artefacts
	rm -rf $(BUILD_DIR)

release: ## Tag and push a release (LEVEL=patch|minor|major)
ifndef LEVEL
	$(error usage: make release LEVEL=patch|minor|major)
endif
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | sed 's/^v//' | cut -d. -f3); \
	case "$(LEVEL)" in \
		patch) PATCH=$$((PATCH + 1)) ;; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0 ;; \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0 ;; \
		*) echo "LEVEL must be patch, minor, or major"; exit 1 ;; \
	esac; \
	TAG="v$$MAJOR.$$MINOR.$$PATCH"; \
	echo "Tagging $$TAG"; \
	git tag -a "$$TAG" -m "Release $$TAG" && \
	echo "Pushing tag $$TAG" && \
	git push origin "$$TAG"
