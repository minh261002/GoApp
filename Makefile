# E-commerce Backend Makefile
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=ecommerce
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER=docker

# Database parameters
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=ecommerce

.PHONY: all build clean test deps run dev migrate worker help

# Default target
all: deps build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/server

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
run: build
	./$(BINARY_NAME)

# Run in development mode
dev:
	$(GOCMD) run ./cmd/server/main.go

# Run with custom port
dev-port:
	$(GOCMD) run ./cmd/server/main.go -port 3000

# Run worker
worker:
	$(GOCMD) run ./cmd/worker/main.go

# Run the email worker
email-worker:
	$(GOCMD) run ./cmd/email-worker/main.go

# Run worker with custom interval
worker-interval:
	$(GOCMD) run ./cmd/worker/main.go -interval 10s

# Run migrations
migrate:
	$(GOCMD) run ./cmd/migrate/main.go

# Run migrations with status check
migrate-status:
	$(GOCMD) run ./cmd/migrate/main.go -action status

# Docker commands
docker-build:
	$(DOCKER) build -t ecommerce-backend .

docker-run:
	$(DOCKER) run -p 8080:8080 ecommerce-backend

# Docker Compose commands
docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-restart:
	$(DOCKER_COMPOSE) restart

# Database commands
db-connect:
	mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME)

db-reset:
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d mysql redis
	sleep 10
	$(MAKE) migrate

# Development setup
setup: deps docker-up
	sleep 10
	$(MAKE) migrate
	@echo "Development environment is ready!"
	@echo "Run 'make dev' to start the server"

# Production setup
prod-setup: deps
	$(MAKE) build-linux
	@echo "Production build completed!"

# Lint and format
lint:
	golangci-lint run

fmt:
	$(GOCMD) fmt ./...

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	@echo "Documentation is available in the documents/ directory"

# Install development tools
install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/swaggo/swag/cmd/swag

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Download dependencies"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run in development mode"
	@echo "  dev-port       - Run with custom port (3000)"
	@echo "  worker         - Run notification worker"
	@echo "  worker-interval- Run worker with custom interval"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-status - Check migration status"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-up      - Start services with Docker Compose"
	@echo "  docker-down    - Stop services with Docker Compose"
	@echo "  docker-logs    - Show Docker Compose logs"
	@echo "  docker-restart - Restart Docker Compose services"
	@echo "  db-connect     - Connect to database"
	@echo "  db-reset       - Reset database (WARNING: destroys data)"
	@echo "  setup          - Setup development environment"
	@echo "  prod-setup     - Setup production build"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  docs           - Generate documentation"
	@echo "  install-tools  - Install development tools"
	@echo "  help           - Show this help message"
