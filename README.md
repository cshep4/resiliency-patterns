# Distributed Systems Patterns in Go

A comprehensive collection of examples and demonstrations showcasing different techniques and design patterns for building resilient distributed systems in Go.

## Overview

This repository contains practical implementations of common distributed systems patterns, focusing on reliability, scalability, and fault tolerance. Each pattern includes working code examples, explanations, and best practices for real-world applications.

## Examples

### Basic Examples
- **[hello-world](examples/hello-world/)** - Basic project structure and Go module organization

### Resilience Patterns (Coming Soon)
- **Circuit Breaker** - Prevent cascading failures by temporarily blocking requests to failing services
- **Retry with Backoff** - Intelligent retry mechanisms with exponential backoff and jitter
- **Timeout Management** - Proper timeout handling for network operations
- **Bulkhead** - Isolate resources to prevent total system failure

### Communication Patterns (Coming Soon)
- **Request-Response** - Synchronous communication patterns
- **Publish-Subscribe** - Asynchronous messaging patterns
- **Message Queues** - Reliable message delivery and processing
- **Event Sourcing** - Event-driven architecture patterns

### Data Patterns (Coming Soon)
- **CQRS** - Command Query Responsibility Segregation
- **Saga Pattern** - Distributed transaction management
- **Event Sourcing** - Storing state as a sequence of events
- **Database Per Service** - Data isolation in microservices

### Observability Patterns (Coming Soon)
- **Health Checks** - Service health monitoring
- **Metrics Collection** - Application and system metrics
- **Distributed Tracing** - Request tracing across services
- **Structured Logging** - Consistent logging practices

## Project Structure

```
├── examples/
│   ├── hello-world/
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   └── Makefile
│   ├── circuit-breaker/
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── Makefile
│   └── retry-pattern/
│       ├── cmd/
│       ├── internal/
│       └── Makefile
├── docs/
├── go.mod
├── go.sum
└── Makefile
```

Each example is self-contained with:
- `cmd/` - Entry points and main applications
- `internal/` - Private implementation code
- `Makefile` - Build and run commands specific to the example

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker (for running examples with external dependencies)
- Make (optional, for convenience commands)

### Quick Start

```bash
git clone https://github.com/your-username/distributed-systems-patterns-go.git
cd distributed-systems-patterns-go
go mod download

# Try the hello-world example
cd examples/hello-world
make run
```

### Running Examples

Each example is self-contained with its own Makefile:

```bash
# Run a specific example
cd examples/hello-world
make run

# Build an example
make build

# Run tests for a specific example
make test

# Run all tests from root
go test ./...

# Run with race detection
go test -race ./...
```

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Resources

- [Microservices Patterns](https://microservices.io/patterns/) by Chris Richardson
- [Building Microservices](https://www.oreilly.com/library/view/building-microservices/9781491950340/) by Sam Newman
- [Designing Data-Intensive Applications](https://dataintensive.net/) by Martin Kleppmann

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Go community for excellent tooling and libraries
- Contributors to open-source distributed systems projects
- Authors of foundational distributed systems research
