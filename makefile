# Load .env file automatically
ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: run migrate migrate-down migrate-reset test lint docker-build docker-up docker-down dev check-containers check-db check-app verify clean frontend-install frontend-dev frontend-build frontend-start

# --- Backend commands --- 
# Run the application locally
run:
	@cd backend && go run ./cmd/server/.

# Run database migrations (via Docker). Added --rm flag to automatically remove 
# migration containers after execution. Leave me alone, I'm learning docker. 
migrate:
	@echo "Running database migrations..."
	@docker compose run --rm migrations

# Roll back the most recent migration
migrate-down:
	@echo "Rolling back migration..."
	@docker compose run --rm migrations sh -c 'goose -dir ./backend/migrations postgres "$$DB_URL" down' # Reset the database

migrate-reset:
	@echo "Resetting database..."
	@docker compose run --rm migrations sh -c 'goose -dir ./backend/migrations postgres "$$DB_URL" reset'
	@docker compose run --rm migrations sh -c 'goose -dir ./backend/migrations postgres "$$DB_URL" up'

migrate-status:
	@echo "Migration status:"
	@docker compose run --rm migrations sh -c 'goose -dir ./backend/migrations postgres "$$DB_URL" status'


## Test and quality commands 

# Run tests
test:
	@echo "Running tests..."
	@cd backend && go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint is not installed. Please install it first."; \
		exit 1; \
	fi


## --- Frontend commands ---
frontend-install:
	@echo "Installing frontend dependencies..."
	@if command -v yarn >/dev/null; then \
		cd frontend && yarn install; \
	else \
		echo "Yarn not found, using npm instead..."; \
		cd frontend && npm install; \
	fi

frontend-dev:
	@echo "Starting frontend development server..."
	@if command -v yarn >/dev/null; then \
		cd frontend && yarn dev; \
	else \
		cd frontend && npm run dev; \
	fi

frontend-build:
	@echo "Building frontend for production..."
	@cd frontend && yarn build

frontend-start:
	@echo "Starting production frontend..."
	@docker compose up -d frontend

# --- Docker commands --- 

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker compose build backend frontend

# Start all containers. Add --wait flag to ensure dependencies are ready
# Again, leave me alone, I'm learning docker. 
docker-up:
	@echo "Starting containers..."
	@docker compose up -d --wait postgres redis backend frontend

# --- Full Development Setup ---
dev-full: docker-up migrate frontend-dev
	@echo "Full development environment ready!"
	@echo "Backend: http://localhost:${BACKEND_PORT}"
	@echo "Frontend: http://localhost:${FRONTEND_PORT}"

# Stop all containers
docker-down:
	@echo "Stopping containers..."
	@docker compose down

## Verification commands 

check-containers:
	@echo "Container status:"
	@docker compose ps

check-db:
	@echo "Testing database connection:"
	@docker compose exec postgres psql -U postgres -d resume_generator -c "\dt"

check-app:
	@echo "Testing application health:"
	@curl -s http://localhost:${PORT}/api/v1/health | jq

verify: check-containers check-db check-app

## Log inspection

logs:
	@docker compose logs --follow

logs-postgres:
	@docker compose logs postgres --follow

logs-app:
	@docker compose logs app --follow

# Dev setup: start Docker and run migrations
dev: docker-up migrate verify
	@echo "Dev environment is up and verified!"

## Deep clean
clean: docker-down
	@echo "Removing volumes..."
	@docker volume prune -f
	@echo "Clean complete!"

# --- Frontend-Specific Clean ---
clean-frontend:
	@echo "Cleaning frontend artifacts..."
	@cd frontend && rm -rf node_modules dist .cache
