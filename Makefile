.PHONY: build test lint fmt clean install example coverage help

# Default target
all: lint test build

# Build the library, CLI, and examples
build:
	@echo "Building..."
	go build ./...
	go build -o bin/imx ./cmd/imx
	go build -o bin/basic ./examples/basic
	go build -o bin/advanced ./examples/advanced

# Run tests with race detector and coverage
test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.out ./...

# Run linter
lint:
	@echo "Running linter..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install CLI to GOPATH/bin
install:
	@echo "Installing CLI..."
	go install ./cmd/imx

# Build and run basic example
example: build
	@echo "Running example..."
	./bin/imx testdata/goldens/jpeg/canon_xmp.jpg

# Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Show help
help:
	@echo "Available targets:"
	@echo "  all       - Run lint, test, and build (default)"
	@echo "  build     - Build the library and CLI"
	@echo "  test      - Run tests with race detector and coverage"
	@echo "  lint      - Run go vet"
	@echo "  fmt       - Format code with go fmt"
	@echo "  clean     - Remove build artifacts"
	@echo "  install   - Install CLI to GOPATH/bin"
	@echo "  example   - Build and run example"
	@echo "  coverage  - Generate HTML coverage report"
	@echo "  help      - Show this help"

