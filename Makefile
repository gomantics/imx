.PHONY: build test test-lib lint fmt clean install coverage coverage-html check help bench bench-all bench-report bench-compare bench-viz

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
	$(GOBUILD) -o $(BIN_DIR)/imx ./cmd/imx
	$(GOBUILD) -o $(BIN_DIR)/basic ./examples/basic
	$(GOBUILD) -o $(BIN_DIR)/advanced ./examples/advanced
	@echo "✓ Build complete"

# Run all tests with race detector
test:
	@echo "Running tests..."
	$(GOTEST) -race $(ALL_PKGS)
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
	@echo "✓ Linting passed"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) $(ALL_PKGS)
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
	$(GOCMD) install ./cmd/imx
	@echo "✓ Installed imx to GOPATH/bin"

# Generate coverage report for library (100% target)
coverage: test-lib
	@echo "Coverage report:"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo ""
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep -v "100.0%" || echo "✓ 100% coverage achieved!"

# Generate HTML coverage report
coverage-html: test-lib
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
	$(GOTEST) -bench=. -benchmem -run=^$$ . ./internal/meta/... ./internal/format/...
	@echo "✓ Benchmarks complete"

# Run benchmarks with detailed output
bench-all:
	@echo "Running all benchmarks with detailed output..."
	$(GOTEST) -bench=. -benchmem -benchtime=3s -run=^$$ . ./internal/meta/... ./internal/format/... | tee bench.txt
	@echo "✓ Benchmarks saved to bench.txt"

# Generate detailed benchmark report
bench-report:
	@echo "Generating benchmark report..."
	@./scripts/bench-report.sh
	@echo "✓ Benchmark report generated"

# Compare benchmarks between commits
bench-compare:
	@echo "Comparing benchmarks between commits..."
	@./scripts/bench-compare.sh $(OLD) $(NEW)
	@echo "Usage: make bench-compare OLD=commit1 NEW=commit2"

# Generate benchmark visualizations across git history
bench-viz:
	@echo "Running benchmarks across git history and generating graphs..."
	@./scripts/bench-viz.py --max-commits $(or $(N),50) --graphs --output out
	@echo "✓ Benchmark visualization complete. Graphs saved to out/"

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
	@echo "  coverage     - Show coverage report (100% target)"
	@echo "  coverage-html- Generate HTML coverage report"
	@echo "  bench        - Run all benchmarks"
	@echo "  bench-all    - Run benchmarks with detailed output"
	@echo "  bench-report - Generate detailed benchmark report"
	@echo "  bench-compare- Compare benchmarks between commits"
	@echo "  bench-viz    - Generate benchmark graphs across git history (N=commits)"
	@echo "  example      - Build and run example"
	@echo "  help         - Show this help"
