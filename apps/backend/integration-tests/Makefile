.PHONY: help test test-real test-db-up test-db-down test-db-reset migrate-up migrate-down

# Default target
help:
	@echo "Eva Health Backend Integration Tests"
	@echo ""
	@echo "Available targets:"
	@echo "  test           - Run integration tests with mock Azure services"
	@echo "  test-real      - Run integration tests with real Azure services"
	@echo "  test-db-up     - Start test database container"
	@echo "  test-db-down   - Stop test database container"
	@echo "  test-db-reset  - Reset test database (stop, remove, start, migrate)"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"

# Database configuration
DB_CONTAINER_NAME := eva-health-test-db
DB_NAME := eva_health_test
DB_USER := postgres
DB_PASSWORD := postgres
DB_PORT := 5432
TEST_DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Start test database
test-db-up:
	@echo "Starting test database..."
	@docker run -d \
		--name $(DB_CONTAINER_NAME) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		postgres:15 || echo "Database container already running"
	@echo "Waiting for database to be ready..."
	@sleep 3
	@docker exec $(DB_CONTAINER_NAME) pg_isready -U $(DB_USER) || sleep 2

# Stop test database
test-db-down:
	@echo "Stopping test database..."
	@docker stop $(DB_CONTAINER_NAME) || true
	@docker rm $(DB_CONTAINER_NAME) || true

# Reset test database
test-db-reset: test-db-down test-db-up migrate-up
	@echo "Test database reset complete"

# Run migrations
migrate-up:
	@echo "Running migrations..."
	@cd .. && migrate -path migrations -database "$(TEST_DATABASE_URL)" up

# Rollback migrations
migrate-down:
	@echo "Rolling back migrations..."
	@cd .. && migrate -path migrations -database "$(TEST_DATABASE_URL)" down

# Run integration tests with mock Azure services
test: test-db-up migrate-up
	@echo "Running integration tests with mock Azure services..."
	@cd .. && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -v ./integration-tests/...

# Run integration tests with real Azure services
test-real: test-db-up migrate-up
	@echo "Running integration tests with real Azure services..."
	@if [ -z "$$AZURE_OPENAI_ENDPOINT" ]; then \
		echo "Error: AZURE_OPENAI_ENDPOINT not set"; \
		exit 1; \
	fi
	@if [ -z "$$AZURE_OPENAI_KEY" ]; then \
		echo "Error: AZURE_OPENAI_KEY not set"; \
		exit 1; \
	fi
	@if [ -z "$$AZURE_SPEECH_KEY" ]; then \
		echo "Error: AZURE_SPEECH_KEY not set"; \
		exit 1; \
	fi
	@cd .. && \
		TEST_DATABASE_URL="$(TEST_DATABASE_URL)" \
		USE_REAL_AZURE=true \
		go test -v ./integration-tests/...

# Clean up everything
clean: test-db-down
	@echo "Cleanup complete"
