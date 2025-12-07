#!/bin/bash
# Script to generate a detailed benchmark report
# Usage: ./scripts/bench-report.sh [output-file]

set -e

OUTPUT=${1:-bench-report.txt}

echo "Running comprehensive benchmarks..."
echo "This may take a few minutes..."
echo ""

# Run benchmarks with extended time for better accuracy
go test -bench=. -benchmem -benchtime=3s -run=^$ ./... | tee "$OUTPUT"

echo ""
echo "=================================================="
echo "Benchmark report saved to: $OUTPUT"
echo "=================================================="
echo ""

# Extract key metrics
echo "Key Performance Metrics:"
echo "------------------------"
echo ""

# Top-level operations
echo "High-Level Operations:"
grep "BenchmarkMetadataFromFile-" "$OUTPUT" | head -3

echo ""
echo "Parser Performance:"
grep "BenchmarkParser_Parse-" "$OUTPUT" | grep -E "(exif|iptc|xmp|icc|jpeg)"

echo ""
echo "Memory Allocations:"
grep "BenchmarkMetadata" "$OUTPUT" | awk '{print $1, $5, $7}'

echo ""
echo "Full report available in: $OUTPUT"
