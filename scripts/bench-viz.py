#!/usr/bin/env python3
"""
Benchmark Visualization Tool for imx

This script runs benchmarks across git commits and generates performance graphs.
It handles parallel execution, error handling, and creates visualizations for:
- ns/op (nanoseconds per operation)
- B/op (bytes allocated per operation)
- allocs/op (allocations per operation)
"""

import argparse
import json
import multiprocessing as mp
import os
import re
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass, asdict
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

try:
    import matplotlib.pyplot as plt
    import matplotlib.dates as mdates
    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False
    print("Warning: matplotlib not found. Install with: pip3 install matplotlib", file=sys.stderr)


@dataclass
class BenchmarkResult:
    """Single benchmark result"""
    name: str
    ns_per_op: float
    bytes_per_op: int
    allocs_per_op: int


@dataclass
class CommitBenchmark:
    """Benchmark results for a single commit"""
    commit_hash: str
    commit_date: datetime
    commit_message: str
    results: List[BenchmarkResult]
    success: bool
    error: Optional[str] = None


class BenchmarkRunner:
    """Runs benchmarks across git commits"""

    def __init__(self, repo_path: Path, max_commits: int = 100, parallel: int = 1):
        self.repo_path = repo_path
        self.max_commits = max_commits
        self.parallel = parallel
        self.original_branch = self._get_current_branch()

    def _get_current_branch(self) -> str:
        """Get the current git branch"""
        result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=self.repo_path,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()

    def _get_commits(self) -> List[Tuple[str, datetime, str]]:
        """Get list of commits with dates and messages"""
        result = subprocess.run(
            ["git", "log", f"-{self.max_commits}", "--format=%H|%at|%s"],
            cwd=self.repo_path,
            capture_output=True,
            text=True,
            check=True
        )

        commits = []
        for line in result.stdout.strip().split("\n"):
            if not line:
                continue
            parts = line.split("|", 2)
            if len(parts) == 3:
                commit_hash, timestamp, message = parts
                commit_date = datetime.fromtimestamp(int(timestamp))
                commits.append((commit_hash, commit_date, message))

        return commits

    def _checkout_commit(self, commit_hash: str) -> bool:
        """Checkout a specific commit"""
        try:
            subprocess.run(
                ["git", "checkout", "-q", commit_hash],
                cwd=self.repo_path,
                capture_output=True,
                check=True
            )
            return True
        except subprocess.CalledProcessError:
            return False

    def _run_benchmark(self, commit_hash: str) -> Tuple[bool, Optional[str], Optional[str]]:
        """Run benchmark for a commit. Returns (success, stdout, stderr)"""
        try:
            result = subprocess.run(
                ["go", "test", "-bench=.", "-benchmem", "-run=^$",
                 ".", "./internal/meta/...", "./internal/format/..."],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                timeout=300  # 5 minute timeout
            )

            # Consider it successful if we got any benchmark output
            if "Benchmark" in result.stdout:
                return True, result.stdout, result.stderr
            else:
                return False, None, result.stderr or "No benchmark output"

        except subprocess.TimeoutExpired:
            return False, None, "Benchmark timeout (>5min)"
        except subprocess.CalledProcessError as e:
            return False, None, f"Benchmark failed: {e}"
        except Exception as e:
            return False, None, f"Unexpected error: {e}"

    def _parse_benchmark_output(self, output: str) -> List[BenchmarkResult]:
        """Parse go test -bench output"""
        results = []

        # Pattern: BenchmarkName-12    1000000    1234 ns/op    5678 B/op    90 allocs/op
        pattern = r'Benchmark(\S+)-\d+\s+\d+\s+([\d.]+)\s+ns/op\s+([\d.]+)\s+B/op\s+([\d.]+)\s+allocs/op'

        for match in re.finditer(pattern, output):
            name = match.group(1)
            ns_per_op = float(match.group(2))
            bytes_per_op = int(float(match.group(3)))
            allocs_per_op = int(float(match.group(4)))

            results.append(BenchmarkResult(
                name=name,
                ns_per_op=ns_per_op,
                bytes_per_op=bytes_per_op,
                allocs_per_op=allocs_per_op
            ))

        return results

    def run_commit_benchmark(self, commit_info: Tuple[str, datetime, str]) -> CommitBenchmark:
        """Run benchmark for a single commit"""
        commit_hash, commit_date, commit_message = commit_info

        print(f"Testing {commit_hash[:8]} - {commit_message[:60]}...", file=sys.stderr)

        if not self._checkout_commit(commit_hash):
            return CommitBenchmark(
                commit_hash=commit_hash,
                commit_date=commit_date,
                commit_message=commit_message,
                results=[],
                success=False,
                error="Failed to checkout commit"
            )

        success, stdout, stderr = self._run_benchmark(commit_hash)

        if success and stdout:
            results = self._parse_benchmark_output(stdout)
            print(f"  ✓ Found {len(results)} benchmarks", file=sys.stderr)
            return CommitBenchmark(
                commit_hash=commit_hash,
                commit_date=commit_date,
                commit_message=commit_message,
                results=results,
                success=True
            )
        else:
            print(f"  ✗ Failed: {stderr[:100] if stderr else 'unknown'}", file=sys.stderr)
            return CommitBenchmark(
                commit_hash=commit_hash,
                commit_date=commit_date,
                commit_message=commit_message,
                results=[],
                success=False,
                error=stderr
            )

    def run(self) -> List[CommitBenchmark]:
        """Run benchmarks across all commits"""
        commits = self._get_commits()
        print(f"Running benchmarks across {len(commits)} commits...", file=sys.stderr)

        results = []

        if self.parallel > 1:
            # Note: Parallel execution with git checkout is tricky
            # For now, run sequentially. TODO: Use worktrees for parallel
            print(f"Note: Parallel execution not yet implemented, running sequentially", file=sys.stderr)

        try:
            for commit_info in commits:
                result = self.run_commit_benchmark(commit_info)
                results.append(result)
        finally:
            # Always return to original branch
            self._checkout_commit(self.original_branch)
            print(f"\nReturned to branch: {self.original_branch}", file=sys.stderr)

        successful = sum(1 for r in results if r.success)
        print(f"\nCompleted: {successful}/{len(results)} commits had successful benchmarks", file=sys.stderr)

        return results


class BenchmarkVisualizer:
    """Creates visualizations from benchmark results"""

    def __init__(self, output_dir: Path):
        self.output_dir = output_dir
        self.output_dir.mkdir(parents=True, exist_ok=True)

        if not HAS_MATPLOTLIB:
            raise ImportError("matplotlib is required for visualization. Install with: pip3 install matplotlib")

    def _organize_by_benchmark(self, commits: List[CommitBenchmark]) -> Dict[str, List[Tuple[datetime, BenchmarkResult]]]:
        """Organize results by benchmark name"""
        by_benchmark = defaultdict(list)

        for commit in commits:
            if not commit.success:
                continue
            for result in commit.results:
                by_benchmark[result.name].append((commit.commit_date, result))

        # Sort by date
        for name in by_benchmark:
            by_benchmark[name].sort(key=lambda x: x[0])

        return dict(by_benchmark)

    def _create_metric_graph(self, benchmark_name: str, data: List[Tuple[datetime, BenchmarkResult]],
                            metric: str, ylabel: str, filename: str):
        """Create a graph for a specific metric"""
        if not data:
            return

        dates = [d for d, _ in data]

        if metric == "ns_per_op":
            values = [r.ns_per_op for _, r in data]
        elif metric == "bytes_per_op":
            values = [r.bytes_per_op for _, r in data]
        elif metric == "allocs_per_op":
            values = [r.allocs_per_op for _, r in data]
        else:
            return

        plt.figure(figsize=(12, 6))
        plt.plot(dates, values, marker='o', linestyle='-', linewidth=2, markersize=4)
        plt.xlabel('Date', fontsize=12)
        plt.ylabel(ylabel, fontsize=12)
        plt.title(f'{benchmark_name} - {ylabel}', fontsize=14, fontweight='bold')
        plt.grid(True, alpha=0.3)

        # Format x-axis
        plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
        plt.gcf().autofmt_xdate()

        # Add latest value annotation
        if values:
            latest_val = values[-1]
            plt.annotate(f'Latest: {latest_val:.2f}',
                        xy=(dates[-1], latest_val),
                        xytext=(10, 10),
                        textcoords='offset points',
                        bbox=dict(boxstyle='round,pad=0.5', fc='yellow', alpha=0.7),
                        arrowprops=dict(arrowstyle='->', connectionstyle='arc3,rad=0'))

        plt.tight_layout()
        plt.savefig(self.output_dir / filename, dpi=150, bbox_inches='tight')
        plt.close()

    def generate_graphs(self, commits: List[CommitBenchmark]):
        """Generate all performance graphs"""
        by_benchmark = self._organize_by_benchmark(commits)

        print(f"\nGenerating graphs for {len(by_benchmark)} benchmarks...", file=sys.stderr)

        for benchmark_name, data in by_benchmark.items():
            if not data:
                continue

            safe_name = benchmark_name.replace("/", "_").replace(" ", "_")

            # Generate graph for each metric
            self._create_metric_graph(
                benchmark_name, data, "ns_per_op",
                "Nanoseconds per Operation",
                f"{safe_name}_ns_per_op.png"
            )

            self._create_metric_graph(
                benchmark_name, data, "bytes_per_op",
                "Bytes Allocated per Operation",
                f"{safe_name}_bytes_per_op.png"
            )

            self._create_metric_graph(
                benchmark_name, data, "allocs_per_op",
                "Allocations per Operation",
                f"{safe_name}_allocs_per_op.png"
            )

            print(f"  Generated graphs for {benchmark_name}", file=sys.stderr)

        print(f"\nGraphs saved to: {self.output_dir}", file=sys.stderr)

    def generate_summary_graphs(self, commits: List[CommitBenchmark]):
        """Generate summary graphs showing multiple benchmarks together"""
        by_benchmark = self._organize_by_benchmark(commits)

        if not by_benchmark:
            return

        # Group benchmarks by category
        categories = {
            "High-Level API": [],
            "Parser - EXIF": [],
            "Parser - IPTC": [],
            "Parser - XMP": [],
            "Parser - ICC": [],
            "Format - JPEG": [],
        }

        for name in by_benchmark.keys():
            if name.startswith("MetadataFrom") or name.startswith("Metadata_") or name in ["SmallFile", "LargeFile", "WithMaxBytes", "WithBufferSize"]:
                categories["High-Level API"].append(name)
            elif name.startswith("Parser_Parse") and "exif" in name.lower():
                categories["Parser - EXIF"].append(name)
            elif name.startswith("Parser_Parse") and "iptc" in name.lower():
                categories["Parser - IPTC"].append(name)
            elif name.startswith("Parser_Parse") and "xmp" in name.lower():
                categories["Parser - XMP"].append(name)
            elif name.startswith("Parser_Parse") and "icc" in name.lower():
                categories["Parser - ICC"].append(name)
            elif name.startswith("Parser_Parse") and "jpeg" in name.lower():
                categories["Format - JPEG"].append(name)

        # Create combined graph for each category
        for category, benchmark_names in categories.items():
            if not benchmark_names:
                continue

            plt.figure(figsize=(14, 7))

            for bench_name in benchmark_names:
                data = by_benchmark[bench_name]
                dates = [d for d, _ in data]
                values = [r.ns_per_op for _, r in data]
                plt.plot(dates, values, marker='o', label=bench_name, linewidth=2, markersize=3)

            plt.xlabel('Date', fontsize=12)
            plt.ylabel('Nanoseconds per Operation', fontsize=12)
            plt.title(f'{category} - Performance Over Time', fontsize=14, fontweight='bold')
            plt.legend(loc='best', fontsize=9)
            plt.grid(True, alpha=0.3)
            plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
            plt.gcf().autofmt_xdate()

            safe_category = category.replace(" ", "_").replace("-", "")
            plt.tight_layout()
            plt.savefig(self.output_dir / f"summary_{safe_category}.png", dpi=150, bbox_inches='tight')
            plt.close()

            print(f"  Generated summary graph for {category}", file=sys.stderr)


def save_json_results(commits: List[CommitBenchmark], output_file: Path):
    """Save benchmark results as JSON"""
    data = []
    for commit in commits:
        commit_data = {
            "commit_hash": commit.commit_hash,
            "commit_date": commit.commit_date.isoformat(),
            "commit_message": commit.commit_message,
            "success": commit.success,
            "error": commit.error,
            "results": [asdict(r) for r in commit.results]
        }
        data.append(commit_data)

    with open(output_file, 'w') as f:
        json.dump(data, f, indent=2)

    print(f"Results saved to: {output_file}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(
        description="Run benchmarks across git history and generate performance graphs",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Run benchmarks on last 50 commits and generate graphs
  %(prog)s -n 50 --graphs

  # Just collect data without graphs
  %(prog)s -n 100 --json-only

  # Generate graphs from existing JSON data
  %(prog)s --from-json bench_results.json --graphs
        """
    )

    parser.add_argument("-n", "--max-commits", type=int, default=100,
                       help="Maximum number of commits to test (default: 100)")
    parser.add_argument("-j", "--parallel", type=int, default=1,
                       help="Number of parallel jobs (default: 1, not yet implemented)")
    parser.add_argument("-o", "--output", type=Path, default=Path("out"),
                       help="Output directory for graphs (default: out/)")
    parser.add_argument("--json", type=Path, default=Path("bench_results.json"),
                       help="JSON file for results (default: bench_results.json)")
    parser.add_argument("--graphs", action="store_true",
                       help="Generate performance graphs")
    parser.add_argument("--json-only", action="store_true",
                       help="Only save JSON, don't generate graphs")
    parser.add_argument("--from-json", type=Path,
                       help="Load results from JSON instead of running benchmarks")

    args = parser.parse_args()

    repo_path = Path.cwd()

    # Load or run benchmarks
    if args.from_json:
        print(f"Loading results from {args.from_json}...", file=sys.stderr)
        with open(args.from_json) as f:
            data = json.load(f)

        commits = []
        for item in data:
            results = [BenchmarkResult(**r) for r in item["results"]]
            commits.append(CommitBenchmark(
                commit_hash=item["commit_hash"],
                commit_date=datetime.fromisoformat(item["commit_date"]),
                commit_message=item["commit_message"],
                results=results,
                success=item["success"],
                error=item.get("error")
            ))
    else:
        runner = BenchmarkRunner(repo_path, args.max_commits, args.parallel)
        commits = runner.run()
        save_json_results(commits, args.json)

    # Generate graphs
    if args.graphs and not args.json_only:
        if not HAS_MATPLOTLIB:
            print("Error: matplotlib is required for graphs. Install with: pip3 install matplotlib", file=sys.stderr)
            sys.exit(1)

        visualizer = BenchmarkVisualizer(args.output)
        visualizer.generate_graphs(commits)
        visualizer.generate_summary_graphs(commits)

        print(f"\n✓ Complete! Graphs available in {args.output}/", file=sys.stderr)

    return 0


if __name__ == "__main__":
    sys.exit(main())
