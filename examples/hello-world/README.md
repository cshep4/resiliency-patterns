# Hello World Example

A simple example demonstrating the basic project structure and Go module organization used throughout this repository.

## Overview

This example shows:
- Basic Go project structure with `cmd/` and `internal/` directories
- Simple package organization and testing
- Command-line flag handling
- Makefile usage for common tasks

## Structure

```
hello-world/
├── cmd/
│   └── main.go          # Application entry point
├── internal/
│   ├── greeter.go       # Business logic
│   └── greeter_test.go  # Unit tests
├── Makefile             # Build and run commands
└── README.md            # This file
```

## Usage

### Using Makefile (Recommended)

```bash
# Run with default settings
make run

# Run with custom name
make run ARGS="-name Alice"

# Run with time
make run ARGS="-name Alice -time"

# Run demo mode
make demo

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Direct Go Commands

```bash
# Run with default settings
go run cmd/main.go

# Run with custom name
go run cmd/main.go -name Alice

# Run with time
go run cmd/main.go -name Alice -time

# Run demo mode
go run cmd/main.go demo

# Run tests
go test ./internal/

# Build
go build -o bin/hello-world cmd/main.go
```

## Example Output

```bash
$ make run
Hello, World! Welcome to Distributed Systems Patterns in Go.

$ make run ARGS="-name Alice -time"
Hello, Alice! The time is 14:30:25. Welcome to Distributed Systems Patterns in Go.

$ make demo
Hello, World! Welcome to Distributed Systems Patterns in Go.

--- Demo Mode ---
Demonstrating different greetings:
- Hello, Alice! Welcome to Distributed Systems Patterns in Go.
- Hello, Bob! Welcome to Distributed Systems Patterns in Go.
- Hello, Charlie! Welcome to Distributed Systems Patterns in Go.

With time:
- Hello, Developer! The time is 14:30:25. Welcome to Distributed Systems Patterns in Go.
```

## Key Concepts

This example demonstrates:

1. **Package Organization**: Separation of concerns with `cmd/` for entry points and `internal/` for implementation
2. **Testing**: Unit tests with table-driven test patterns
3. **Command-line Interface**: Using the `flag` package for CLI arguments
4. **Build Automation**: Makefile for common development tasks
5. **Go Modules**: Proper import paths using the module name

This structure serves as the foundation for all other examples in this repository.
