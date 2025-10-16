.PHONY: run build test clean install dev example help

# Default target
all: build

# Application name
APP_NAME=link-preview-api
BUILD_DIR=./bin
MAIN_FILE=main.go

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Install dependencies
install:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	go mod tidy
	go mod download
	@echo "$(GREEN)Dependencies installed successfully!$(NC)"

# Run the application in development mode
dev: install
	@echo "$(BLUE)Starting development server...$(NC)"
	@echo "$(YELLOW)Server will be available at: http://localhost:5465$(NC)"
	@echo "$(YELLOW)API Documentation: http://localhost:5465/$(NC)"
	@echo "$(YELLOW)Health Check: http://localhost:5465/health$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop the server$(NC)"
	GIN_MODE=debug go run $(MAIN_FILE)

# Run the application
run: build
	@echo "$(BLUE)Starting production server...$(NC)"
	$(BUILD_DIR)/$(APP_NAME)

# Build the application
build: install
	@echo "$(BLUE)Building application...$(NC)"
	mkdir -p $(BUILD_DIR)
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-s -w" \
		-o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-s -w" \
		-o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	GO111MODULE=on CGO_ENABLED=0 go build \
		-ldflags "-s -w" \
		-o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Build completed successfully!$(NC)"
	@echo "$(YELLOW)Binaries available in $(BUILD_DIR)/$(NC)"

# Run tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)Tests completed!$(NC)"

# Run example client (requires server to be running)
example:
	@echo "$(BLUE)Running API examples...$(NC)"
	@echo "$(YELLOW)Make sure the server is running with 'make dev' first!$(NC)"
	go run -ldflags "-X main.runExamples=true" $(MAIN_FILE) example_test.go

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

# Lint code
lint:
	@echo "$(BLUE)Linting code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi
	@echo "$(GREEN)Linting completed!$(NC)"

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean
	@echo "$(GREEN)Clean completed!$(NC)"

# Show project structure
struct:
	@echo "$(BLUE)Project Structure:$(NC)"
	tree -I 'bin|.git' || find . -type f -name '*.go' -o -name '*.md' -o -name 'Makefile' | grep -v bin | sort

# Test API endpoints with curl
test-api:
	@echo "$(BLUE)Testing API endpoints...$(NC)"
	@echo "$(YELLOW)Make sure the server is running first!$(NC)"
	@echo "\n$(BLUE)1. Testing health check...$(NC)"
	curl -s http://localhost:5465/health | jq . || curl -s http://localhost:5465/health
	@echo "\n\n$(BLUE)2. Testing GitHub preview...$(NC)"
	curl -s -X POST http://localhost:5465/preview \
		-H "Content-Type: application/json" \
		-d '{"url": "https://github.com"}' | jq . || \
	curl -s -X POST http://localhost:5465/preview \
		-H "Content-Type: application/json" \
		-d '{"url": "https://github.com"}'
	@echo "\n\n$(BLUE)3. Testing API documentation...$(NC)"
	curl -s http://localhost:5465/ | jq . || curl -s http://localhost:5465/
	@echo "\n$(GREEN)API testing completed!$(NC)"

# Docker build
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)Docker image built successfully!$(NC)"

# Docker run
docker-run: docker-build
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 5465:5465 --rm $(APP_NAME):latest

# Show help
help:
	@echo "$(BLUE)Available targets:$(NC)"
	@echo "  $(GREEN)install$(NC)      - Install dependencies"
	@echo "  $(GREEN)dev$(NC)          - Run in development mode with hot reload"
	@echo "  $(GREEN)run$(NC)          - Run the built application"
	@echo "  $(GREEN)build$(NC)        - Build the application for multiple platforms"
	@echo "  $(GREEN)test$(NC)         - Run tests"
	@echo "  $(GREEN)example$(NC)      - Run example client (server must be running)"
	@echo "  $(GREEN)test-api$(NC)     - Test API endpoints with curl"
	@echo "  $(GREEN)fmt$(NC)          - Format code"
	@echo "  $(GREEN)lint$(NC)         - Lint code"
	@echo "  $(GREEN)clean$(NC)        - Clean build artifacts"
	@echo "  $(GREEN)struct$(NC)       - Show project structure"
	@echo "  $(GREEN)docker-build$(NC) - Build Docker image"
	@echo "  $(GREEN)docker-run$(NC)   - Run Docker container"
	@echo "  $(GREEN)help$(NC)         - Show this help message"
	@echo ""
	@echo "$(YELLOW)Quick start:$(NC)"
	@echo "  1. $(GREEN)make dev$(NC)     - Start development server"
	@echo "  2. $(GREEN)make test-api$(NC) - Test the API (in another terminal)"
	@echo "  3. $(GREEN)make build$(NC)   - Build for production"

