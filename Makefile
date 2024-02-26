.ONESHELL:
.DEFAULT_GOAL := help

# allow user specific optional overrides
-include Makefile.overrides

export

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: up
up: ## runs everything
	@docker-compose up --build --force-recreate

.PHONY: down
down: ## stop all systems
	@docker-compose down --remove-orphans

.PHONY: rm
rm: ## remove everything
	@docker-compose down --volumes --remove-orphans

.PHONY: build
build: ## build and generate api binary
	@go build -o out/api ./...

.PHONY: run
run: ## run api locally
	@go run ./...

.PHONY: test
test: ## run local tests
	@go test ./cmd/... -v -race

.PHONY: deps
deps: ## download dependencies
	@go mod download

.PHONY: dev-db
dev-db: ## run dev env
	@docker-compose -f ./deployments/docker-compose-db.yml up --force-recreate --build
