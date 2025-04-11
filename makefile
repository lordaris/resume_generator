.PHONY: run migrate test lint

# Run the application
run:
	@go run ./cmd/server/main.go

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@goose -dir ./migrations postgres "$(DB_URL)" up

# Roll back the most recent migration
migrate-down:
	@echo "Rolling back the most recent migration..."
	@goose -dir ./migrations postgres "$(DB_URL)" down

# Reset the database by rolling back all migrations and reapplying them
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
