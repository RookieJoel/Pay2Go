.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down

# Variables
APP_NAME=pay2go
CMD_DIR=./cmd/api
BUILD_DIR=./bin
MIGRATE_DIR=./migrations

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_DIR)/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@migrate -path $(MIGRATE_DIR) -database "postgres://postgres:postgres@localhost:5433/pay2go?sslmode=disable" up
	@echo "Migrations complete"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path $(MIGRATE_DIR) -database "postgres://postgres:postgres@localhost:5433/pay2go?sslmode=disable" down
	@echo "Rollback complete"

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(NAME)

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Docker containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "Docker containers stopped"

docker-build: ## Build Docker image for the application
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .
	@echo "Docker image built"

dev: docker-up ## Start development environment
	@echo "Waiting for database to be ready..."
	@sleep 3
	@make migrate-up
	@make run

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

security-scan: ## Run Snyk security scan
	@echo "Running Snyk security scan..."
	@snyk test --all-projects
	@echo "Security scan complete"

all: clean deps build test ## Clean, download deps, build and test
