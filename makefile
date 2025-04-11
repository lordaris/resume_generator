
# Load .env file automatically
ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: run migrate migrate-down migrate-reset test lint docker-build docker-up docker-down dev

# Run the application locally
run:
	@go run ./cmd/server/.

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@goose -dir ./migrations postgres "$(DB_URL)" up

# Roll back the most recent migration
migrate-down:
	@echo "Rolling back the most recent migration..."
	@goose -dir ./migrations postgres "$(DB_URL)" down

# Reset the database
migrate-reset:
	@echo "Resetting database..."
	@goose -dir ./migrations postgres "$(DB_URL)" reset
	@goose -dir ./migrations postgres "$(DB_URL)" up

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

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker compose build

# Start all containers
docker-up:
	@echo "Starting containers..."
	@docker compose up -d

# Stop all containers
docker-down:
	@echo "Stopping containers..."
	@docker compose down

# Dev setup: start Docker and run migrations
dev: docker-up migrate
	@echo "Dev environment is up!"

