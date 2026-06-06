.PHONY: build run test test-coverage migrate-up migrate-down seed clean docker-up docker-down dev-infra dev-db dev test-docker test-compose-e2e keygen docker-build docker-buildx help

# Variables
BINARY_NAME=server
MAIN_PATH=./cmd/server
KEYGEN_PATH=./cmd/keygen
SEED_PATH=./cmd/seed
DATABASE_URL?=postgres://keyles:keyles_dev_password@localhost:5432/keyles?sslmode=disable
MIGRATIONS_PATH=./migrations
DOCKER_COMPOSE ?= docker compose -f ../docker-compose.yml
CONTAINER_ENGINE ?= podman

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
	@echo "  dev-infra      - Start local Docker dependencies"
	@echo "  dev-db         - Run migrations and seed data"
	@echo "  dev            - Prepare local dependencies and run server"
	@echo "  test-docker    - Run integration tests with Docker dependencies"
	@echo "  test-compose-e2e - Run the live OAuth browser matrix against the Compose stack"
	@echo "  docker-build  - Build Docker image for native architecture"
	@echo "  docker-buildx - Build multi-arch image (amd64+arm64) and push"

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
	@go test -coverpkg=./... -coverprofile=coverage.out ./...
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
	@$(DOCKER_COMPOSE) up -d || { echo "Docker Compose failed. If your setup uses the legacy binary, retry with: make DOCKER_COMPOSE='docker-compose -f ../docker-compose.yml' docker-up"; exit 1; }
	@echo "✓ Docker services started"

# Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@$(DOCKER_COMPOSE) down || { echo "Docker Compose failed. If your setup uses the legacy binary, retry with: make DOCKER_COMPOSE='docker-compose -f ../docker-compose.yml' docker-down"; exit 1; }
	@echo "✓ Docker services stopped"

# Development workflow shortcuts
dev-infra: docker-up
	@echo "✓ Development infrastructure ready"

dev-db: migrate-up seed
	@echo "✓ Development database ready"

dev: dev-infra dev-db
	@echo "✓ Development environment ready"
	@$(MAKE) run

# Run tests in Docker
test-docker: docker-up
	@echo "Running tests in Docker environment..."
	@trap '$(MAKE) docker-down' EXIT; go test -v ./tests/integration/...

# Run the seeded OAuth browser-flow matrix against the active Compose stack.
test-compose-e2e:
	@echo "Running live OAuth browser-flow matrix..."
	@KEYLES_COMPOSE_E2E=1 CONTAINER_ENGINE=$(CONTAINER_ENGINE) go test -count=1 -run TestDockerComposeOAuthBrowserFlow -v ./tests/integration


# Build Docker image for native architecture
docker-build:
	@echo "Building Docker image for native architecture..."
	@docker build -t keyles-backend:latest .
	@echo "✓ Docker image built: keyles-backend:latest"

# Build multi-architecture Docker image (requires docker buildx + push to registry)
docker-buildx:
	@echo "Creating/using buildx builder..."
	@docker buildx inspect multiarch-builder >/dev/null 2>&1 || docker buildx create --name multiarch-builder --use
	@docker buildx use multiarch-builder
	@echo "Building multi-arch image (amd64+arm64)..."
	@docker buildx build --platform linux/amd64,linux/arm64 -t keyles-backend:latest --push .
	@echo "✓ Multi-arch image pushed: keyles-backend:latest"

# Generate random secrets for .env file
generate-secrets:
	@echo "Generating random secrets..."
	@echo "JWT_SECRET=$$(openssl rand -base64 48)" > .env.secrets
	@echo "DB_PASSWORD=$$(openssl rand -base64 32)" >> .env.secrets
	@echo "REDIS_PASSWORD=$$(openssl rand -base64 32)" >> .env.secrets
	@echo "BREVO_API_KEY=your_brevo_api_key_here" >> .env.secrets
	@echo "✓ Secrets written to .env.secrets"
	@echo "Copy these values into your .env file"
