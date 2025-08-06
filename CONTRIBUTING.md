# Contributing Guidelines

Thank you for your interest in contributing to Distributed Systems Patterns in Go! This document provides guidelines for contributing to the project.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker to report bugs or request features
- Provide clear, detailed descriptions with steps to reproduce
- Include Go version, OS, and relevant environment details

### Submitting Changes

1. **Fork the repository** and create a feature branch
2. **Follow Go conventions** - use `gofmt`, `go vet`, and `golint`
3. **Write tests** for new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass** before submitting

### Pattern Implementation Guidelines

When adding new patterns:

1. **Create a dedicated directory** under the appropriate category
2. **Include a README.md** explaining the pattern, when to use it, and trade-offs
3. **Provide working examples** with clear, commented code
4. **Add comprehensive tests** covering normal and edge cases
5. **Include benchmarks** where performance is relevant

### Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Keep functions focused and small
- Add comments for exported functions and complex logic
- Use Go modules for dependency management

### Testing

- Write unit tests for all new code
- Include integration tests for complex patterns
- Use table-driven tests where appropriate
- Aim for high test coverage
- Include examples in tests using `Example` functions

### Documentation

- Update README.md if adding new patterns
- Include inline documentation for exported functions
- Provide usage examples in pattern directories
- Keep documentation concise but comprehensive

## Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/distributed-systems-patterns-go.git
cd distributed-systems-patterns-go

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Pull Request Process

1. Update documentation and tests
2. Ensure CI passes
3. Request review from maintainers
4. Address feedback promptly
5. Squash commits before merge

## Questions?

Feel free to open an issue for questions about contributing or implementation details.
