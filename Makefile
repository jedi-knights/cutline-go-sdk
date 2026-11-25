SHELL := /usr/bin/env bash

MODULE   := github.com/jedi-knights/cutline-go-sdk
VERSION  ?= dev
LDFLAGS  := -ldflags "-s -w -X $(MODULE).Version=$(VERSION)"

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run unit tests with race detector and coverage
	go test -race -coverprofile=coverage.out ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: vuln
vuln: ## Run govulncheck
	govulncheck ./...

.PHONY: tidy
tidy: ## Sync go.sum with go.mod
	go mod tidy

.PHONY: examples
examples: ## Run every example against the live service (needs CUTLINE_API_KEY)
	@for ex in ./examples/*/; do \
		echo "===== $$ex ====="; \
		go run "$$ex" || exit 1; \
	done

.PHONY: build-examples
build-examples: ## Compile every example without running
	@for ex in ./examples/*/; do \
		echo "===== $$ex ====="; \
		go build $(LDFLAGS) -o /dev/null "$$ex" || exit 1; \
	done

.PHONY: clean
clean: ## Remove build artifacts
	rm -f coverage.out
