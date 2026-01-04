.PHONY: build run test test-coverage clean deps verify lint lint-fix

# Binary name
BINARY_NAME=k8s-tui

# Build the application
build:
	go build -o $(BINARY_NAME) .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Verify dependencies
verify:
	go mod verify

# Run linter
lint:
	golangci-lint run

# Run linter and fix auto-fixable issues
lint-fix:
	golangci-lint run --fix
