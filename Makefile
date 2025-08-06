.PHONY: help test test-race fmt vet lint clean build examples

# Default target
help:
	@echo "Available targets:"
	@echo "  test       - Run all tests"
	@echo "  test-race  - Run tests with race detection"
	@echo "  fmt        - Format all Go code"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Run golint (requires golint to be installed)"
	@echo "  clean      - Clean build artifacts"
	@echo "  build      - Build all examples"
	@echo "  examples   - Run all examples"

# Run tests
test:
	go test ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint code (requires golint)
lint:
	golint ./...

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out

# Build all examples
build:
	@echo "Building examples..."
	@find patterns -name "*.go" -not -name "*_test.go" | xargs -I {} dirname {} | sort -u | xargs -I {} sh -c 'echo "Building {}" && cd {} && go build -o /dev/null .'

# Run examples (where applicable)
examples:
	@echo "Running examples..."
	@find examples -name "main.go" | xargs -I {} dirname {} | xargs -I {} sh -c 'echo "Running example in {}" && cd {} && timeout 5s go run . || true'

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install development dependencies
deps:
	go mod download
	go install golang.org/x/lint/golint@latest

# Check everything
check: fmt vet test-race
	@echo "All checks passed!"
