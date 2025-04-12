# Load .env file automatically
ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: start stop restart build dev backend-container frontend-container \
	db-only migrate migrate-down migrate-reset migrate-status run-backend-local \
	run-frontend-local frontend-install frontend-build test lint check-containers \
	check-db check-app verify clean help

.DEFAULT_GOAL := help

# --- Primary commands ---
start: ## Start all containers and show URLs
	@echo "Starting all containers..."
	@docker compose up -d
	@echo "Services available at:"
	@echo " - Backend: http://localhost:${BACKEND_PORT}"
	@echo " - Frontend: http://localhost:${FRONTEND_PORT}"

stop: ## Stop all running containers
	@echo "Stopping all containers..."
	@docker compose down

restart: ## Restart all services (stop + start)
	@echo "Restarting services..."
	@docker compose restart

build: ## Rebuild all Docker images after changes
	@echo "Building all images..."
	@docker compose build

# --- Development workflows ---
dev: ## Full setup: start services, run migrations, and verify everything
dev: start migrate verify
	@echo "Starting development environment..."
	@echo "Development environment ready!"

backend-container: ## Start only backend + database services
	@docker compose up -d postgres redis backend

frontend-container: ## Start only frontend service
	@docker compose up -d frontend

# --- Database operations ---
db-only: ## Start only database and Redis
	@echo "Starting database services..."
	@docker compose up -d postgres redis

migrate: ## Apply database migrations
	@echo "Running database migrations..."
	@docker compose run --rm migrations

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" down'

migrate-reset: ## Reset database to clean state (all migrations down then up)
	@echo "Resetting database..."
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" reset'
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" up'

migrate-status: ## Show current migration status
	@echo "Migration status:"
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" status'

# --- Local development ---
run-backend-local: ## Run backend locally (requires running database)
	@echo "Starting local backend server..."
	@cd backend && go run ./cmd/server/.

run-frontend-local: frontend-install ## Run frontend locally with dev server
	@echo "Starting local frontend dev server..."
	@if command -v yarn >/dev/null; then \
	  cd frontend && yarn dev; \
	else \
	  cd frontend && npm run dev; \
	fi

# --- Frontend operations ---
frontend-install: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	@if command -v yarn >/dev/null; then \
	 cd frontend && yarn install; \
	else \
	 echo "Yarn not found, using npm instead..."; \
	 cd frontend && npm install; \
	fi

frontend-build: ## Build production frontend assets
	@echo "Building frontend for production..."
	@cd frontend && yarn build

# --- Quality control ---
test: ## Run backend tests
	@echo "Running tests..."
	@cd backend && go test -v ./...

lint: ## Run code linter
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
	 golangci-lint run ./...; \
	else \
	 echo "golangci-lint is not installed. Please install it first."; \
	 exit 1; \
	fi

# --- System checks ---
check-containers: ## Show container statuses
	@echo "Container status:"
	@docker compose ps

check-db: ## Verify database connection
	@echo "Testing database connection:"
	@docker compose exec postgres psql -U postgres -d resume_generator -c "\dt"

check-app: ## Check application health status
	@echo "Testing application health:"
	@curl -s http://localhost:${PORT}/api/v1/health | jq

verify: ## Verify system health (containers + DB + app)
verify: check-containers check-db check-app

# --- Maintenance ---
clean: ## Remove all containers, volumes, and build artifacts
	@echo "Cleaning system..."
	@docker compose down -v
	@echo "Removing frontend artifacts..."
	@cd frontend && rm -rf node_modules dist .cache
	@echo "Cleanup complete!"

# --- Documentation ---
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
