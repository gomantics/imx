#!/usr/bin/env python3
"""
Comprehensive Benchmark Tool for imx

Runs benchmarks, generates human-readable reports, and creates historical performance graphs.
Integrates directly with `make bench` - no separate commands needed.

Usage:
    make bench              # Run benchmarks and show current results
    make bench N=50         # Run benchmarks + generate graphs for last 50 commits
"""

import argparse
import re
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

try:
    import matplotlib.pyplot as plt
    import matplotlib.dates as mdates
    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False


@dataclass
class BenchmarkMetrics:
    """Comprehensive benchmark metrics"""
    name: str
    package: str  # Go package name
    iterations: int
    ns_per_op: float
    bytes_per_op: int
    allocs_per_op: int


@dataclass
class CommitBenchmark:
    """Benchmark results for a commit"""
    commit_hash: str
    commit_date: datetime
    commit_message: str
    metrics: List[BenchmarkMetrics]
    success: bool
    error: Optional[str] = None


class BenchmarkRunner:
    """Runs benchmarks and collects metrics"""

    def __init__(self, repo_path: Path):
        self.repo_path = repo_path

    def run_current_benchmarks(self) -> Tuple[bool, List[BenchmarkMetrics], str]:
        """Run benchmarks on current commit"""
        try:
            result = subprocess.run(
                ["go", "test", "-bench=.", "-benchmem", "-benchtime=2s", "-run=^$",
                 ".", "./internal/meta/...", "./internal/format/..."],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                timeout=300
            )

            if "Benchmark" in result.stdout:
                metrics = self._parse_benchmark_output(result.stdout)
                return True, metrics, result.stdout
            else:
                return False, [], result.stderr or "No benchmark output"

        except subprocess.TimeoutExpired:
            return False, [], "Benchmark timeout (>5min)"
        except Exception as e:
            return False, [], f"Error: {e}"

    def _parse_benchmark_output(self, output: str) -> List[BenchmarkMetrics]:
        """Parse Go benchmark output"""
        metrics = []

        # Pattern: BenchmarkName-12    1000000    1234 ns/op    5678 B/op    90 allocs/op
        pattern = r'Benchmark(\S+)-\d+\s+(\d+)\s+([\d.]+)\s+ns/op\s+([\d.]+)\s+B/op\s+([\d.]+)\s+allocs/op'

        # Parse benchmarks line by line to track current package
        current_pkg = "unknown"
        for line in output.split("\n"):
            if line.startswith("pkg: "):
                current_pkg = line.split("pkg: ")[1].strip()
            elif line.startswith("Benchmark"):
                match = re.match(pattern, line)
                if match:
                    name = match.group(1)
                    iterations = int(match.group(2))
                    ns_per_op = float(match.group(3))
                    bytes_per_op = int(float(match.group(4)))
                    allocs_per_op = int(float(match.group(5)))

                    metrics.append(BenchmarkMetrics(
                        name=name,
                        package=current_pkg,
                        iterations=iterations,
                        ns_per_op=ns_per_op,
                        bytes_per_op=bytes_per_op,
                        allocs_per_op=allocs_per_op
                    ))

        return metrics


class ReportFormatter:
    """Formats benchmark results as human-readable reports"""

    @staticmethod
    def format_number(n: float) -> str:
        """Format number with K/M/B suffixes"""
        if n == 0:
            return "-"

        if n >= 1_000_000_000:
            return f"{n/1_000_000_000:.2f}B"
        elif n >= 1_000_000:
            return f"{n/1_000_000:.2f}M"
        elif n >= 1_000:
            return f"{n/1_000:.2f}K"
        elif n >= 1:
            return f"{n:.2f}"
        else:
            return f"{n:.2f}"

    @staticmethod
    def format_time(ns: float) -> str:
        """Format time with appropriate unit (ns, μs, ms, s)"""
        if ns == 0:
            return "-"

        if ns >= 1_000_000_000:  # >= 1 second
            return f"{ns/1_000_000_000:.2f}s"
        elif ns >= 1_000_000:  # >= 1 millisecond
            return f"{ns/1_000_000:.2f}ms"
        elif ns >= 1_000:  # >= 1 microsecond
            return f"{ns/1_000:.2f}µs"
        else:  # nanoseconds
            return f"{ns:.2f}ns"

    @staticmethod
    def format_bytes(b: int) -> str:
        """Format bytes with B/KB/MB suffixes"""
        if b == 0:
            return "-"

        if b >= 1024*1024:
            return f"{b/(1024*1024):.2f}MB"
        elif b >= 1024:
            return f"{b/1024:.2f}KB"
        else:
            return f"{b}B"

    @staticmethod
    def format_current_results(metrics: List[BenchmarkMetrics], output: str) -> str:
        """Format current benchmark results"""
        if not metrics:
            return "No benchmarks found"

        lines = []
        lines.append("=" * 100)
        lines.append("BENCHMARK RESULTS")
        lines.append("=" * 100)
        lines.append("")

        # Group by category using package field
        categories = defaultdict(list)

        for m in metrics:
            # Categorize based on package
            if "internal/meta/exif" in m.package:
                categories["EXIF Parser"].append(m)
            elif "internal/meta/iptc" in m.package:
                categories["IPTC Parser"].append(m)
            elif "internal/meta/xmp" in m.package:
                categories["XMP Parser"].append(m)
            elif "internal/meta/icc" in m.package:
                categories["ICC Parser"].append(m)
            elif "internal/format/jpeg" in m.package:
                categories["JPEG Format"].append(m)
            elif m.package.endswith("/imx") or m.package == "github.com/gomantics/imx":
                categories["High-Level API"].append(m)
            else:
                categories["Other"].append(m)

        # Print each category
        for category, cat_metrics in sorted(categories.items()):
            if not cat_metrics:
                continue

            lines.append(f"\n{category}")
            lines.append("-" * 100)
            lines.append(f"{'Benchmark':<50} {'Iterations':>12} {'Latency/op':>12} {'B/op':>12} {'allocs/op':>12}")
            lines.append("-" * 100)

            for m in cat_metrics:
                iters_str = ReportFormatter.format_number(m.iterations)
                latency_str = ReportFormatter.format_time(m.ns_per_op)
                bytes_str = ReportFormatter.format_bytes(m.bytes_per_op)
                allocs_str = ReportFormatter.format_number(m.allocs_per_op) if m.allocs_per_op > 0 else "-"

                lines.append(f"{m.name:<50} {iters_str:>12} {latency_str:>12} {bytes_str:>12} {allocs_str:>12}")

        lines.append("")
        lines.append("=" * 100)
        lines.append("SUMMARY")
        lines.append("=" * 100)

        # Calculate summary stats
        total_benchmarks = len(metrics)
        avg_ns = sum(m.ns_per_op for m in metrics) / total_benchmarks if total_benchmarks > 0 else 0
        total_allocs = sum(m.bytes_per_op for m in metrics)
        fastest = min(metrics, key=lambda m: m.ns_per_op) if metrics else None
        slowest = max(metrics, key=lambda m: m.ns_per_op) if metrics else None

        lines.append(f"Total Benchmarks: {total_benchmarks}")
        lines.append(f"Average Latency: {ReportFormatter.format_time(avg_ns)}/op")
        lines.append(f"Total Memory Allocated: {ReportFormatter.format_bytes(total_allocs)} across all benchmarks")
        if fastest:
            lines.append(f"Fastest: {fastest.name} ({ReportFormatter.format_time(fastest.ns_per_op)}/op)")
        if slowest:
            lines.append(f"Slowest: {slowest.name} ({ReportFormatter.format_time(slowest.ns_per_op)}/op)")
        lines.append("")

        return "\n".join(lines)


class HistoricalRunner:
    """Runs benchmarks across git history"""

    def __init__(self, repo_path: Path, max_commits: int):
        self.repo_path = repo_path
        self.max_commits = max_commits
        self.original_branch = self._get_current_branch()

    def _get_current_branch(self) -> str:
        """Get current git branch"""
        result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=self.repo_path,
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()

    def _get_commits(self) -> List[Tuple[str, datetime, str]]:
        """Get list of commits"""
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
        """Checkout a commit"""
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

    def run(self) -> List[CommitBenchmark]:
        """Run benchmarks across commits"""
        commits = self._get_commits()
        print(f"Running benchmarks across {len(commits)} commits...", file=sys.stderr)
        print("This may take a while...\n", file=sys.stderr)

        results = []
        runner = BenchmarkRunner(self.repo_path)

        try:
            for i, (commit_hash, commit_date, commit_message) in enumerate(commits, 1):
                short_hash = commit_hash[:8]
                print(f"[{i}/{len(commits)}] {short_hash} - {commit_message[:50]}...", end=" ", file=sys.stderr, flush=True)

                if not self._checkout_commit(commit_hash):
                    print("✗ checkout failed", file=sys.stderr)
                    results.append(CommitBenchmark(
                        commit_hash=commit_hash,
                        commit_date=commit_date,
                        commit_message=commit_message,
                        metrics=[],
                        success=False,
                        error="Checkout failed"
                    ))
                    continue

                success, metrics, error = runner.run_current_benchmarks()

                if success:
                    print(f"✓ {len(metrics)} benchmarks", file=sys.stderr)
                    results.append(CommitBenchmark(
                        commit_hash=commit_hash,
                        commit_date=commit_date,
                        commit_message=commit_message,
                        metrics=metrics,
                        success=True
                    ))
                else:
                    print(f"✗ {error[:50]}", file=sys.stderr)
                    results.append(CommitBenchmark(
                        commit_hash=commit_hash,
                        commit_date=commit_date,
                        commit_message=commit_message,
                        metrics=[],
                        success=False,
                        error=error
                    ))

        finally:
            self._checkout_commit(self.original_branch)
            print(f"\nReturned to branch: {self.original_branch}", file=sys.stderr)

        successful = sum(1 for r in results if r.success)
        print(f"Completed: {successful}/{len(results)} commits had successful benchmarks\n", file=sys.stderr)

        return results


class GraphGenerator:
    """Generates performance graphs"""

    def __init__(self, output_dir: Path):
        self.output_dir = output_dir
        self.output_dir.mkdir(parents=True, exist_ok=True)

        if not HAS_MATPLOTLIB:
            raise ImportError("matplotlib required. Install with: pip3 install matplotlib")

    def _organize_by_benchmark(self, commits: List[CommitBenchmark]) -> Dict[str, List[Tuple[datetime, BenchmarkMetrics]]]:
        """Organize results by benchmark name"""
        by_benchmark = defaultdict(list)

        for commit in commits:
            if not commit.success:
                continue
            for metric in commit.metrics:
                by_benchmark[metric.name].append((commit.commit_date, metric))

        # Sort by date
        for name in by_benchmark:
            by_benchmark[name].sort(key=lambda x: x[0])

        return dict(by_benchmark)

    def generate(self, commits: List[CommitBenchmark]):
        """Generate all performance graphs"""
        by_benchmark = self._organize_by_benchmark(commits)

        if not by_benchmark:
            print("No benchmark data to visualize", file=sys.stderr)
            return

        # Generate metric graphs (all benchmarks combined)
        self._generate_metric_graph(by_benchmark, "ops_per_sec", "Operations per Second", "ops_per_sec.png")
        self._generate_metric_graph(by_benchmark, "ns_per_op", "Nanoseconds per Operation", "ns_per_op.png")
        self._generate_metric_graph(by_benchmark, "bytes_per_op", "Bytes Allocated per Operation", "bytes_per_op.png")
        self._generate_metric_graph(by_benchmark, "allocs_per_op", "Allocations per Operation", "allocs_per_op.png")
        self._generate_metric_graph(by_benchmark, "mb_per_sec", "Throughput (MB/s)", "throughput.png")

        print(f"\n✓ Graphs saved to {self.output_dir}/", file=sys.stderr)

    def _generate_metric_graph(self, by_benchmark: Dict, metric_name: str, ylabel: str, filename: str):
        """Generate a single metric graph with all benchmarks"""
        plt.figure(figsize=(16, 10))

        # Group benchmarks by category for better colors
        categories = {
            "API": [],
            "EXIF": [],
            "IPTC": [],
            "XMP": [],
            "ICC": [],
            "JPEG": [],
        }

        for name in by_benchmark.keys():
            if name.startswith("MetadataFrom") or name.startswith("Metadata_"):
                categories["API"].append(name)
            elif "exif" in name.lower() or "IFD" in name:
                categories["EXIF"].append(name)
            elif "iptc" in name.lower():
                categories["IPTC"].append(name)
            elif "xmp" in name.lower():
                categories["XMP"].append(name)
            elif "icc" in name.lower():
                categories["ICC"].append(name)
            else:
                categories["JPEG"].append(name)

        color_map = {
            "API": "tab:blue",
            "EXIF": "tab:orange",
            "IPTC": "tab:green",
            "XMP": "tab:red",
            "ICC": "tab:purple",
            "JPEG": "tab:brown",
        }

        for category, bench_names in categories.items():
            for bench_name in bench_names:
                data = by_benchmark.get(bench_name, [])
                if not data:
                    continue

                dates = [d for d, _ in data]

                if metric_name == "ops_per_sec":
                    values = [m.ops_per_sec for _, m in data]
                elif metric_name == "ns_per_op":
                    values = [m.ns_per_op for _, m in data]
                elif metric_name == "bytes_per_op":
                    values = [m.bytes_per_op for _, m in data]
                elif metric_name == "allocs_per_op":
                    values = [m.allocs_per_op for _, m in data]
                elif metric_name == "mb_per_sec":
                    values = [m.mb_per_sec for _, m in data]
                else:
                    continue

                # Skip if all zeros (for mb_per_sec)
                if all(v == 0 for v in values):
                    continue

                plt.plot(dates, values, marker='o', label=f"{bench_name} ({category})",
                        linewidth=2, markersize=4, color=color_map[category], alpha=0.7)

        plt.xlabel('Date', fontsize=14, fontweight='bold')
        plt.ylabel(ylabel, fontsize=14, fontweight='bold')
        plt.title(f'Performance History - {ylabel}', fontsize=16, fontweight='bold')
        plt.legend(loc='best', fontsize=8, ncol=2)
        plt.grid(True, alpha=0.3, linestyle='--')
        plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
        plt.gcf().autofmt_xdate()
        plt.tight_layout()
        plt.savefig(self.output_dir / filename, dpi=150, bbox_inches='tight')
        plt.close()

        print(f"  ✓ Generated {filename}", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser(
        description="Comprehensive benchmark tool - run benchmarks and generate reports",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    parser.add_argument("-n", "--history", type=int, metavar="N",
                       help="Generate historical graphs for last N commits")
    parser.add_argument("-o", "--output", type=Path, default=Path("out"),
                       help="Output directory for graphs (default: out/)")

    args = parser.parse_args()

    repo_path = Path.cwd()
    runner = BenchmarkRunner(repo_path)

    # Always run current benchmarks first
    print("Running benchmarks on current commit...\n", file=sys.stderr)
    success, metrics, output = runner.run_current_benchmarks()

    if success:
        # Print human-readable report
        report = ReportFormatter.format_current_results(metrics, output)
        print(report)
    else:
        print(f"Error running benchmarks: {output}", file=sys.stderr)
        return 1

    # If history requested, run across commits and generate graphs
    if args.history:
        if not HAS_MATPLOTLIB:
            print("\nWarning: matplotlib not installed. Skipping graph generation.", file=sys.stderr)
            print("Install with: pip3 install matplotlib", file=sys.stderr)
            return 0

        print(f"\n{'='*100}", file=sys.stderr)
        print(f"HISTORICAL ANALYSIS - Last {args.history} commits", file=sys.stderr)
        print(f"{'='*100}\n", file=sys.stderr)

        historical = HistoricalRunner(repo_path, args.history)
        commits = historical.run()

        generator = GraphGenerator(args.output)
        generator.generate(commits)

    return 0


if __name__ == "__main__":
    sys.exit(main())
