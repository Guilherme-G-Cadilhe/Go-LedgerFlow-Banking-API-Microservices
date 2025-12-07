# Makefile
.PHONY: help setup up down logs sqlc migrate test test-integration lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install dependencies and tools
	go mod download
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

up: ## Start all services in background
	docker-compose up -d

down: ## Stop all services
	docker-compose down

logs: ## Follow logs
	docker-compose logs -f

sqlc: ## Generate Go code from SQL queries
	sqlc generate

migrate-up: ## Run migrations up
	migrate -path migrations -database "postgresql://ledger:secret123@localhost:5432/ledgerflow?sslmode=disable" up

test: ## Run unit tests
	go test -v -race -cover ./...

test-integration: ## Run integration tests (requires Docker)
	go test -v -tags=integration ./tests/integration/...