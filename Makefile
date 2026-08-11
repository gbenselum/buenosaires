# Buenos Aires development tasks.
# Usage: make <target>  (e.g. make lint, make test)

.PHONY: help fmt lint sec test check hooks install-hooks

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

fmt: ## Format all Go source files
	gofmt -w .

lint: ## Run golangci-lint
	golangci-lint run ./...

sec: ## Run gosec security scanner
	gosec ./...

test: ## Run the Go test suite
	go test ./...

check: lint sec test ## Run all local checks (lint + security + tests)

install-hooks: ## Install pre-commit hooks
	pre-commit install

hooks: ## Run all pre-commit hooks against the full tree
	pre-commit run --all-files
