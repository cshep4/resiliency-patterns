# Resiliency Patterns

A collection of examples and demonstrations showcasing different techniques and design patterns for building resilient distributed systems in Go.

## Overview

This repository contains practical implementations of common resiliency patterns discussed in the [Bit Summit 2025](https://bit-summit.com/#agenda) talk "Building Ultra-Resilient Systems in Go".

## Patterns

### Ensuring High Availability
- **Active/Active Deployments** - Deploy multiple active instances to eliminate single points of failure
- **Service Discovery & Registry** - Dynamic service registration and discovery for resilient architectures
- **[Leader Election & Coordination](/high-availability/leader-election/)** - Coordinate distributed systems with leader election patterns

### Isolate Failures
- **Bulkhead Pattern** - Isolate resources to prevent cascading failures across system components
- **CQRS** - Command Query Responsibility Segregation for separating read and write operations

### Mitigating External Dependency Risk
- **Retries & Timeouts** - Intelligent retry mechanisms with exponential backoff and proper timeout handling
- **[Data Caching](/external-dependency-risk/cache/)** - Cache strategies to reduce load on external services and improve performance
- **[Circuit Breaker](/external-dependency-risk/circuit-breaker/)** - Prevent cascading failures with circuit breakers

Each example is self-contained with:
- `cmd/` - Entry points and main applications
- `internal/` - Private implementation code
- `Makefile` - Build and run commands specific to the example

## Getting Started

### Prerequisites

- Go 1.25 or later
- Make (optional, for convenience commands)

### Running Examples

Each example is self-contained with its own Makefile:

```bash
# Run an example
cd examples/external-dependency-risk/circuit-breaker
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
- **Mitigating External Dependency Risk**: Patterns that mitigate risks from external service dependencies