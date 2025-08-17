# Resiliency Patterns

A comprehensive collection of examples and demonstrations showcasing different techniques and design patterns for building resilient distributed systems in Go.

## Overview

This repository contains practical implementations of common distributed systems patterns, focusing on reliability, scalability, and fault tolerance. Each pattern includes working code examples, explanations, and best practices for real-world applications.

## Examples

### Basic Examples
- **[hello-world](examples/hello-world/)** - Basic project structure and Go module organization

### Ensuring High Availability
- **Active/Active Deployments** - Deploy multiple active instances to eliminate single points of failure
- **Service Discovery & Registry** - Dynamic service registration and discovery for resilient architectures
- **[Leader Election & Coordination](/high-availability/leader-election/)** - Coordinate distributed systems with leader election patterns

### Isolate Failures
- **Bulkhead Pattern** - Isolate resources to prevent cascading failures across system components
- **CQRS** - Command Query Responsibility Segregation for separating read and write operations

### Mitigating External Dependency Risk
- **Retries & Timeouts** - Intelligent retry mechanisms with exponential backoff and proper timeout handling
- **Data Caching** - Cache strategies to reduce dependency on external services and improve performance
- **Circuit Breaker & Fallback** - Prevent cascading failures with circuit breakers and graceful fallback mechanisms

## Project Structure

```
├── high-availability/
│   ├── active-active/              # Active/Active Deployments
│   ├── service-discovery/          # Service Discovery & Registry
│   └── leader-election/            # Leader Election & Coordination
├── isolate-failures/
│   ├── bulkhead/                   # Bulkhead Pattern
│   └── cqrs/                       # CQRS
├── external-dependency-risk/
│   ├── retries-timeouts/           # Retries & Timeouts
│   ├── data-caching/               # Data Caching
│   └── circuit-breaker/            # Circuit Breaker
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

- Go 1.23 or later
- Docker (for running examples with external dependencies)
- Make (optional, for convenience commands)

### Quick Start

```bash
git clone https://github.com/cshep4/resiliency-patterns.git
cd resiliency-patterns
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

# Run a pattern example
cd examples/external-dependency-risk/circuit-breaker-fallback
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

### Pattern Categories

Examples are organized by their primary purpose:

- **High Availability**: Patterns that ensure system uptime and availability
- **Isolate Failures**: Patterns that prevent failures from cascading through the system
- **External Dependency Risk**: Patterns that mitigate risks from external service dependencies

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
