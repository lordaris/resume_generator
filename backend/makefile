# Load .env file automatically
ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: run migrate migrate-down migrate-reset test lint docker-build docker-up docker-down dev check-containers check-db check-app verify clean

# Run the application locally
run:
	@go run ./cmd/server/.

# Run database migrations (via Docker). Added --rm flag to automatically remove 
# migration containers after execution. Leave me alone, I'm learning docker. 
migrate:
	@echo "Running database migrations..."
	@docker compose run --rm migrations

# Roll back the most recent migration
migrate-down:
	@echo "Rolling back migration..."
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" down' # Reset the database

migrate-reset:
	@echo "Resetting database..."
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" reset'
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" up'

migrate-status:
	@echo "Migration status:"
	@docker compose run --rm migrations sh -c 'goose -dir ./migrations postgres "$$DB_URL" status'


## Test and quality commands 

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint is not installed. Please install it first."; \
		exit 1; \
	fi


## Docker operations 

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker compose build

# Start all containers. Add --wait flag to ensure dependencies are ready
# Again, leave me alone, I'm learning docker. 
docker-up:
	@echo "Starting containers..."
	@docker compose up -d --wait postgres redis

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
	@curl -s http://localhost:8080/api/v1/health | jq

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
