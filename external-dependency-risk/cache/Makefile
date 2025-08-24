.PHONY: help run build mocks test clean

# Default target
help:
	@echo "Available targets:"
	@echo "  run    - Run the cache example"
	@echo "  build  - Build the binary"
	@echo "  test   - Run tests"
	@echo "  test-race - Run tests with race detection"
	@echo "  mocks  - Generate mocks"
	@echo "  clean  - Clean build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make run"
	@echo "  make test"

# Binary name
BINARY_NAME=cache

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

# Run tests with coverage
test-cover:
	go test -cover ./...

# Generate mocks
mocks:
	go generate ./...

# Clean build artifacts
clean:
	rm -rf bin/