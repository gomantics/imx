.PHONY: build test test-lib lint fmt clean install coverage coverage-html check help bench

# Go settings
GOCMD = go
GOTEST = $(GOCMD) test
GOBUILD = $(GOCMD) build
GOVET = $(GOCMD) vet
GOFMT = $(GOCMD) fmt

# Directories
BIN_DIR = bin
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

# Library packages (excluding cmd and examples)
LIB_PKGS = ./internal/... .

# All packages
ALL_PKGS = ./...

# Default target
all: check

# Full check: lint, test, build
check: lint test build
	@echo "✓ All checks passed"

# Build the library, CLI, and examples
build:
	@echo "Building..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(ALL_PKGS)
	cd cmd/imx && $(GOBUILD) -o ../../$(BIN_DIR)/imx .
	$(GOBUILD) -o $(BIN_DIR)/basic ./examples/basic
	$(GOBUILD) -o $(BIN_DIR)/advanced ./examples/advanced
	@echo "✓ Build complete"

# Run all tests with race detector
test:
	@echo "Running tests..."
	$(GOTEST) -race $(ALL_PKGS)
	cd cmd/imx && $(GOTEST) -race ./...
	@echo "✓ All tests passed"

# Run library tests only with coverage (excludes cmd and examples)
test-lib:
	@echo "Running library tests with coverage..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(LIB_PKGS)
	@echo "✓ Library tests passed"

# Run linter (go vet)
lint:
	@echo "Running linter..."
	$(GOVET) $(ALL_PKGS)
	cd cmd/imx && $(GOVET) ./...
	@echo "✓ Linting passed"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) $(ALL_PKGS)
	cd cmd/imx && $(GOFMT) ./...
	@echo "✓ Formatting complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "✓ Clean complete"

# Install CLI to GOPATH/bin
install:
	@echo "Installing CLI..."
	cd cmd/imx && $(GOCMD) install .
	@echo "✓ Installed imx to GOPATH/bin"

# Generate coverage report for all packages (library + CLI)
coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(LIB_PKGS)
	@cd cmd/imx && $(GOTEST) -coverprofile=cli-coverage.tmp -covermode=atomic ./...
	@tail -n +2 cmd/imx/cli-coverage.tmp >> $(COVERAGE_FILE)
	@rm -f cmd/imx/cli-coverage.tmp
	@echo ""
	@$(GOTEST) -cover $(LIB_PKGS) 2>&1 | grep "coverage:" | grep -v "no test files"
	@cd cmd/imx && $(GOTEST) -cover ./... 2>&1 | grep "coverage:" | grep -v "no test files"

# Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✓ Coverage report: $(COVERAGE_HTML)"

# Run basic example
example: build
	@echo "Running example..."
	./$(BIN_DIR)/imx testdata/goldens/jpeg/google_iptc.jpg

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -benchtime=1s $(ALL_PKGS)
	cd cmd/imx && $(GOTEST) -bench=. -benchmem -benchtime=1s ./...

# Show help
help:
	@echo "imx - Image Metadata Extraction Library"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Run check (default)"
	@echo "  check        - Run lint, test, and build"
	@echo "  build        - Build library, CLI, and examples"
	@echo "  test         - Run all tests with race detector"
	@echo "  test-lib     - Run library tests with coverage"
	@echo "  lint         - Run go vet"
	@echo "  fmt          - Format code with go fmt"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install CLI to GOPATH/bin"
	@echo "  coverage     - Run tests with coverage (library + CLI)"
	@echo "  coverage-html- Generate HTML coverage report (library + CLI)"
	@echo "  bench        - Run benchmarks"
	@echo "  example      - Build and run example"
	@echo "  help         - Show this help"
