.ONESHELL:
.DEFAULT_GOAL := help

# allow user specific optional overrides
-include Makefile.overrides

export

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## build and generate api binary
	@go build -o out/api ./cmd/...

.PHONY: run
run: ## run api locally
	@go run cmd/api/main.go

.PHONY: test
test: ## run local tests
	@go test ./cmd/... -v -race

.PHONY: deps
deps: ## download dependencies
	@go mod download
