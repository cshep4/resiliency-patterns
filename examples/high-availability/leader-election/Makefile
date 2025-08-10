.PHONY: help run demo build test clean fmt vet run-node1 run-node2 run-node3 demo-parallel

# Default target
help:
	@echo "Available targets:"
	@echo "  run           - Run single instance (use NODE_ID=... to specify node)"
	@echo "  demo          - Run demo with instructions"
	@echo "  demo-parallel - Run 3 instances in parallel (requires tmux or separate terminals)"
	@echo "  run-node1     - Run node-1 instance"
	@echo "  run-node2     - Run node-2 instance" 
	@echo "  run-node3     - Run node-3 instance"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo ""
	@echo "Examples:"
	@echo "  make run NODE_ID=node-1"
	@echo "  make demo"
	@echo "  make demo-parallel  # Requires tmux"

# Binary name
BINARY_NAME=leader-election
BUILD_DIR=bin

# Default node ID
NODE_ID ?= node-1

# Run single instance
run:
	go run cmd/main.go -node-id=$(NODE_ID)

# Run demo with instructions
demo:
	@echo "🚀 Leader Election Demo"
	@echo "======================"
	@echo ""
	@echo "This will start a single instance. To see leader election in action:"
	@echo "1. Open 3 separate terminals"
	@echo "2. In terminal 1: make run-node1"
	@echo "3. In terminal 2: make run-node2" 
	@echo "4. In terminal 3: make run-node3"
	@echo ""
	@echo "You'll see one node become leader and the others remain followers."
	@echo "Try stopping the leader (Ctrl+C) to see another node take over!"
	@echo ""
	@echo "Starting node-1 in 3 seconds..."
	@sleep 3
	@make run-node1

# Run specific nodes
run-node1:
	go run cmd/main.go -node-id=node-1

run-node2:
	go run cmd/main.go -node-id=node-2

run-node3:
	go run cmd/main.go -node-id=node-3

# Run 3 instances in parallel using tmux (if available)
demo-parallel:
	@if command -v tmux >/dev/null 2>&1; then \
		echo "🚀 Starting 3 instances in tmux session..."; \
		tmux new-session -d -s leader-election-demo; \
		tmux split-window -h; \
		tmux split-window -v; \
		tmux select-pane -t 0; \
		tmux split-window -v; \
		tmux send-keys -t 0 'make run-node1' Enter; \
		tmux send-keys -t 1 'make run-node2' Enter; \
		tmux send-keys -t 2 'make run-node3' Enter; \
		tmux select-pane -t 3; \
		tmux send-keys -t 3 'echo "Leader Election Demo Running!"' Enter; \
		tmux send-keys -t 3 'echo "Watch the logs above to see leader election in action"' Enter; \
		tmux send-keys -t 3 'echo "Press Ctrl+C in any pane to stop that instance"' Enter; \
		tmux send-keys -t 3 'echo "Use: tmux kill-session -t leader-election-demo to stop all"' Enter; \
		tmux attach-session -t leader-election-demo; \
	else \
		echo "❌ tmux not found. Please run instances manually:"; \
		echo "Terminal 1: make run-node1"; \
		echo "Terminal 2: make run-node2"; \
		echo "Terminal 3: make run-node3"; \
	fi

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	go test ./internal/...

# Run tests with coverage
test-coverage:
	go test -cover ./internal/...

# Run tests with race detection
test-race:
	go test -race ./internal/...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts and lock files
clean:
	rm -rf $(BUILD_DIR)
	rm -f /tmp/leader-election-demo.lock
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run all checks
check: fmt vet test-race
	@echo "All checks passed!"

# Development workflow
dev: clean fmt vet test build
	@echo "Development build complete!"

# Kill any running tmux session
kill-demo:
	@if tmux has-session -t leader-election-demo 2>/dev/null; then \
		tmux kill-session -t leader-election-demo; \
		echo "Killed leader-election-demo tmux session"; \
	else \
		echo "No leader-election-demo tmux session found"; \
	fi
