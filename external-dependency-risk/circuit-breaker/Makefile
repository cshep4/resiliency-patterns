.PHONY: help run build test clean fmt vet

# Default target
help:
	@echo "Available targets:"
	@echo "  run    - Run the circuit breaker example"
	@echo "  build  - Build the binary"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts"
	@echo "  fmt    - Format code"
	@echo "  vet    - Run go vet"
	@echo ""
	@echo "Examples:"
	@echo "  make run"
	@echo "  make test"

# Binary name
BINARY_NAME=circuit-breaker

# Build the binary
build:
	go build -o bin/$(BINARY_NAME) ./cmd

# Run the example
run:
	go run ./cmd

# Run tests
test:
	go test ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run all checks
check: fmt vet test
