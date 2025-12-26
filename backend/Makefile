.PHONY: build run test test-coverage migrate-up migrate-down seed clean docker-up docker-down keygen help

# Variables
BINARY_NAME=server
MAIN_PATH=./cmd/server
KEYGEN_PATH=./cmd/keygen
SEED_PATH=./cmd/seed
DATABASE_URL?=postgres://keyles:keyles_dev_password@localhost:5432/keyles?sslmode=disable
MIGRATIONS_PATH=./migrations

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  run            - Run the server"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  keygen         - Generate RSA keypair for JWT signing"
	@echo "  seed           - Seed database with test data"
	@echo "  clean          - Remove build artifacts"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go
	@echo "✓ Build complete: bin/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting server..."
	@go run $(MAIN_PATH)/main.go

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out

# Run database migrations up
migrate-up:
	@echo "Running migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
	@echo "✓ Migrations complete"

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1
	@echo "✓ Rollback complete"

# Generate RSA keypair for JWT signing
keygen:
	@echo "Generating RSA keypair..."
	@go run $(KEYGEN_PATH)/main.go
	@echo "✓ Keypair generated"

# Seed database with test data
seed:
	@echo "Seeding database..."
	@go run $(SEED_PATH)/main.go
	@echo "✓ Seed complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

# Start Docker services
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "✓ Docker services started"

# Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "✓ Docker services stopped"

# Development workflow shortcuts
dev: docker-up migrate-up seed
	@echo "✓ Development environment ready"
	@$(MAKE) run

# Run tests in Docker
test-docker: docker-up
	@echo "Running tests in Docker environment..."
	@go test -v ./tests/integration/...
	@$(MAKE) docker-down
