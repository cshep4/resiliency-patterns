.PHONY: help run demo build test clean fmt vet

# Default target
help:
	@echo "Available targets:"
	@echo "  run        - Run the application (use ARGS='...' for arguments)"
	@echo "  demo       - Run in demo mode"
	@echo "  build      - Build the binary"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo ""
	@echo "Examples:"
	@echo "  make run ARGS='-name Alice'"
	@echo "  make run ARGS='-name Bob -time'"

# Binary name
BINARY_NAME=hello-world
BUILD_DIR=bin

# Run the application
run:
	go run cmd/main.go $(ARGS)

# Run in demo mode
demo:
	go run cmd/main.go demo

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	go test ./internal/

# Run tests with coverage
test-coverage:
	go test -cover ./internal/

# Run tests with race detection
test-race:
	go test -race ./internal/

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Install dependencies (if any)
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt vet test-race
	@echo "All checks passed!"

# Development workflow
dev: clean fmt vet test build
	@echo "Development build complete!"
