.PHONY: help build build-worker run run-worker run-prod test vet tidy infra-up infra-down clean

ACTIVE_ENV ?= dev

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the server binary
	go build -o bin/expense-server ./cmd

build-worker: ## Build the WhatsApp Kafka consumer binary
	go build -tags worker -o bin/whatsapp-worker ./cmd

run: ## Run the server (ACTIVE_ENV defaults to dev)
	ACTIVE_ENV=$(ACTIVE_ENV) go run ./cmd

run-worker: ## Run the WhatsApp worker (requires Kafka; set MSGQUEUE_ENABLED=true)
	ACTIVE_ENV=$(ACTIVE_ENV) MSGQUEUE_ENABLED=true go run -tags worker ./cmd

run-prod: ## Run prod-mode locally (release mode) against brew PG/Redis + db expense_service_prod
	ACTIVE_ENV=prod \
	STORE_HOST=localhost STORE_PORT=5432 STORE_USER=auth STORE_PASSWORD=auth \
	STORE_NAME=expense_service_prod STORE_SSLMODE=disable \
	CACHE_HOST=localhost CACHE_PORT=6379 CACHE_TLS=false CACHE_DB=0 \
	TOKEN_PUBLICKEYPATH=$$PWD/assets/keys/jwt_public.pem \
	go run ./cmd

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy modules
	go mod tidy

infra-up: ## Start local Postgres (Redis is shared with auth-service)
	docker compose up -d

infra-down: ## Stop local infrastructure
	docker compose down

clean: ## Remove build artifacts
	rm -rf bin
