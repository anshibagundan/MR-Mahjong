.PHONY: build run test lint clean swagger-gen swagger-check

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mahjong-backend
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/main.go

# Build the application
build:
	@echo "Building application..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build completed: $(BINARY_PATH)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Lint the code
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Generate swagger documentation
swagger-gen:
	@echo "Generating swagger documentation..."
	@which swag > /dev/null || (echo "swag not installed. Installing..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g $(MAIN_PATH) -o ./swagger/gen
	@echo "Swagger documentation generated in ./swagger/gen"

# Check swagger documentation
swagger-check:
	@echo "Checking swagger documentation..."
	@which swag > /dev/null || (echo "swag not installed. Please run 'make swagger-gen' first" && exit 1)
	swag fmt -g $(MAIN_PATH)
	@echo "Swagger documentation check completed"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Development environment setup completed"

# Run application with hot reload (requires air)
dev:
	@echo "Running application with hot reload..."
	@which air > /dev/null || (echo "air not installed. Installing..." && go install github.com/cosmtrek/air@latest)
	air

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

# Run application directly
run-local:
	@echo "Building and running application locally..."
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	$(BINARY_PATH)

# Docker compose up
run:
	@echo "Starting services with docker-compose..."
	docker compose up -d

# Docker compose down
down:
	@echo "Stopping services with docker-compose..."
	docker compose down

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  lint           - Lint the code"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  swagger-gen    - Generate swagger documentation"
	@echo "  swagger-check  - Check swagger documentation"
	@echo "  dev-setup      - Setup development environment"
	@echo "  dev            - Run with hot reload"
	@echo "  docker-build   - Build Docker image"
	@echo "  run            - Start services with docker-compose"
	@echo "  down		    - Stop services with docker-compose"
	@echo "  help           - Show this help message"