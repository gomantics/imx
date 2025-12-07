#!/bin/bash
# Script to compare benchmark results between commits
# Usage: ./scripts/bench-compare.sh [old-commit] [new-commit]

set -e

OLD_COMMIT=${1:-HEAD~1}
NEW_COMMIT=${2:-HEAD}

echo "Comparing benchmarks between $OLD_COMMIT and $NEW_COMMIT"
echo "=================================================="
echo ""

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Save current state
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
git stash push -m "bench-compare temp stash" || true

# Run benchmarks for old commit
echo "Running benchmarks for $OLD_COMMIT..."
git checkout $OLD_COMMIT 2>/dev/null || git checkout -
go test -bench=. -benchmem -run=^$ ./... > "$TEMP_DIR/old.txt" 2>&1

# Run benchmarks for new commit
echo "Running benchmarks for $NEW_COMMIT..."
git checkout $NEW_COMMIT 2>/dev/null || git checkout -
go test -bench=. -benchmem -run=^$ ./... > "$TEMP_DIR/new.txt" 2>&1

# Restore original state
git checkout $CURRENT_BRANCH 2>/dev/null
git stash pop 2>/dev/null || true

# Check if benchstat is installed
if ! command -v benchstat &> /dev/null; then
    echo ""
    echo "Installing benchstat..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Compare results
echo ""
echo "Comparison Results:"
echo "=================================================="
benchstat "$TEMP_DIR/old.txt" "$TEMP_DIR/new.txt"

echo ""
echo "Done! Raw results saved in:"
echo "  Old: $TEMP_DIR/old.txt"
echo "  New: $TEMP_DIR/new.txt"
