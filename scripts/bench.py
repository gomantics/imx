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
        """Run benchmarks and return metrics (for historical analysis)"""
        try:
            result = subprocess.run(
                ["go", "test", "-bench=.", "-benchmem", "-benchtime=1s", "-run=^$",
                 ".", "./internal/meta/...", "./internal/format/..."],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                timeout=300
            )

            if result.returncode == 0:
                metrics = self._parse_benchmark_output(result.stdout)
                if metrics:
                    return True, metrics, result.stdout
                else:
                    return False, [], "No benchmark output"
            else:
                return False, [], result.stderr or "Benchmark failed"

        except subprocess.TimeoutExpired:
            return False, [], "Benchmark timeout (>5min)"
        except Exception as e:
            return False, [], f"Error: {e}"

    def run_current_benchmarks_streaming(self) -> Tuple[bool, List[BenchmarkMetrics], str]:
        """Run benchmarks with streaming formatted output"""
        try:
            process = subprocess.Popen(
                ["go", "test", "-bench=.", "-benchmem", "-benchtime=1s", "-run=^$",
                ".", "./internal/meta/...", "./internal/format/..."],
                cwd=self.repo_path,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1
            )

            output_lines = []
            current_pkg = "unknown"
            metrics = []
            pattern = r'Benchmark(\S+)-\d+\s+(\d+)\s+([\d.]+)\s+ns/op\s+([\d.]+)\s+B/op\s+([\d.]+)\s+allocs/op'

            # Track current category for streaming display
            current_category = None
            category_started = False

            # Stream output line by line
            for line in process.stdout:
                output_lines.append(line)

                # Parse package changes
                if line.startswith("pkg: "):
                    current_pkg = line.split("pkg: ")[1].strip()

                    # Determine category from package
                    if "internal/meta/exif" in current_pkg:
                        new_category = "EXIF Parser"
                    elif "internal/meta/iptc" in current_pkg:
                        new_category = "IPTC Parser"
                    elif "internal/meta/xmp" in current_pkg:
                        new_category = "XMP Parser"
                    elif "internal/meta/icc" in current_pkg:
                        new_category = "ICC Parser"
                    elif "internal/format/jpeg" in current_pkg:
                        new_category = "JPEG Format"
                    elif current_pkg.endswith("/imx") or current_pkg == "github.com/gomantics/imx":
                        new_category = "High-Level API"
                    else:
                        new_category = "Other"

                    # Print category header if changed
                    if new_category != current_category:
                        if category_started:
                            print(flush=True)  # Blank line between categories
                        current_category = new_category
                        category_started = True
                        print(f"\n{current_category}", flush=True)
                        print("-" * 100, flush=True)
                        print(f"{'Benchmark':<50} {'Iterations':>12} {'Latency/op':>12} {'B/op':>12} {'allocs/op':>12}", flush=True)
                        print("-" * 100, flush=True)

                # Parse and display benchmarks as they complete
                elif line.startswith("Benchmark"):
                    match = re.match(pattern, line)
                    if match:
                        name = match.group(1)
                        iterations = int(match.group(2))
                        ns_per_op = float(match.group(3))
                        bytes_per_op = int(float(match.group(4)))
                        allocs_per_op = int(float(match.group(5)))

                        metric = BenchmarkMetrics(
                            name=name,
                            package=current_pkg,
                            iterations=iterations,
                            ns_per_op=ns_per_op,
                            bytes_per_op=bytes_per_op,
                            allocs_per_op=allocs_per_op
                        )
                        metrics.append(metric)

                        # Format and print the result immediately
                        iters_str = ReportFormatter.format_number(iterations)
                        latency_str = ReportFormatter.format_time(ns_per_op)
                        bytes_str = ReportFormatter.format_bytes(bytes_per_op)
                        allocs_str = ReportFormatter.format_number(allocs_per_op) if allocs_per_op > 0 else "-"

                        print(f"{name:<50} {iters_str:>12} {latency_str:>12} {bytes_str:>12} {allocs_str:>12}", flush=True)

            process.wait(timeout=300)
            output = ''.join(output_lines)

            # Print summary
            if metrics:
                print(flush=True)
                print("=" * 100, flush=True)
                print("SUMMARY", flush=True)
                print("=" * 100, flush=True)

                total_benchmarks = len(metrics)

                # Performance breakdown by category
                by_category = {}
                for m in metrics:
                    if "internal/meta/exif" in m.package:
                        cat = "EXIF Parser"
                    elif "internal/meta/iptc" in m.package:
                        cat = "IPTC Parser"
                    elif "internal/meta/xmp" in m.package:
                        cat = "XMP Parser"
                    elif "internal/meta/icc" in m.package:
                        cat = "ICC Parser"
                    elif "internal/format/jpeg" in m.package:
                        cat = "JPEG Format"
                    elif m.package.endswith("/imx") or m.package == "github.com/gomantics/imx":
                        cat = "High-Level API"
                    else:
                        cat = "Other"

                    if cat not in by_category:
                        by_category[cat] = []
                    by_category[cat].append(m)

                print(f"Total: {total_benchmarks} benchmarks across {len(by_category)} categories", flush=True)
                print(flush=True)

                # Category performance
                for cat in sorted(by_category.keys()):
                    cat_metrics = by_category[cat]
                    avg_latency = sum(m.ns_per_op for m in cat_metrics) / len(cat_metrics)
                    avg_mem = sum(m.bytes_per_op for m in cat_metrics) / len(cat_metrics)
                    total_iters = sum(m.iterations for m in cat_metrics)

                    print(f"  {cat:<20} {len(cat_metrics):>2} benchmarks  "
                          f"avg: {ReportFormatter.format_time(avg_latency)}/op  "
                          f"mem: {ReportFormatter.format_bytes(int(avg_mem))}/op  "
                          f"iters: {ReportFormatter.format_number(total_iters)}", flush=True)

                print(flush=True)

            if process.returncode == 0 and metrics:
                return True, metrics, output
            else:
                stderr = process.stderr.read() if process.stderr else ""
                return False, [], stderr or "No benchmark output"

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

    def _has_uncommitted_changes(self) -> bool:
        """Check if there are uncommitted changes"""
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            cwd=self.repo_path,
            capture_output=True,
            text=True
        )
        return bool(result.stdout.strip())

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
        # Check for uncommitted changes
        if self._has_uncommitted_changes():
            print("\nError: Uncommitted changes detected in working directory.", file=sys.stderr)
            print("Please commit or stash your changes before running historical benchmarks:", file=sys.stderr)
            print("  git add .", file=sys.stderr)
            print("  git commit -m 'your message'", file=sys.stderr)
            print("OR", file=sys.stderr)
            print("  git stash", file=sys.stderr)
            sys.exit(1)

        commits = self._get_commits()
        print(f"Running benchmarks across {len(commits)} commits...", file=sys.stderr)
        print("This may take a while...\n", file=sys.stderr)

        results = []
        runner = BenchmarkRunner(self.repo_path)

        try:
            for i, (commit_hash, commit_date, commit_message) in enumerate(commits, 1):
                short_hash = commit_hash[:8]
                commit_date_str = commit_date.strftime("%Y-%m-%d %H:%M:%S")

                print(f"\n[{i}/{len(commits)}] Processing commit {short_hash}", file=sys.stderr)
                print(f"  Date: {commit_date_str}", file=sys.stderr)
                print(f"  Message: {commit_message}", file=sys.stderr)
                print(f"  Checking out...", end=" ", file=sys.stderr, flush=True)

                if not self._checkout_commit(commit_hash):
                    print("✗ FAILED", file=sys.stderr)
                    print(f"  Error: Could not checkout commit", file=sys.stderr)
                    results.append(CommitBenchmark(
                        commit_hash=commit_hash,
                        commit_date=commit_date,
                        commit_message=commit_message,
                        metrics=[],
                        success=False,
                        error="Checkout failed"
                    ))
                    continue

                print("✓", file=sys.stderr)

                # Clean build cache for this commit
                print(f"  Cleaning build cache...", end=" ", file=sys.stderr, flush=True)
                subprocess.run(["go", "clean", "-cache"], cwd=self.repo_path, capture_output=True)
                print("✓", file=sys.stderr)

                print(f"  Running benchmarks:\n", file=sys.stderr)

                success, metrics, error = runner.run_current_benchmarks_streaming()

                if success:
                    print(f"\n  Status: SUCCESS - {len(metrics)} benchmarks completed", file=sys.stderr)
                    results.append(CommitBenchmark(
                        commit_hash=commit_hash,
                        commit_date=commit_date,
                        commit_message=commit_message,
                        metrics=metrics,
                        success=True
                    ))
                else:
                    print(f"\n  Status: FAILED", file=sys.stderr)
                    print(f"  Error: {error[:100]}", file=sys.stderr)
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

    def _is_top_level_benchmark(self, name: str) -> bool:
        """Check if benchmark is a top-level function worth graphing"""
        # Include high-level API benchmarks
        if name.startswith("MetadataFrom"):
            return True
        if name.startswith("Metadata_"):
            return True

        # Include main parser benchmarks (Parser_Parse, Parser_New)
        # Exclude detailed benchmarks like Parser_ParseSegment, Parser_ParseValue, etc.
        if name.startswith("Parser_"):
            return name in ["Parser_Parse", "Parser_New"]

        return False

    def _organize_by_benchmark(self, commits: List[CommitBenchmark]) -> Dict[str, List[Tuple[int, BenchmarkMetrics]]]:
        """Organize results by benchmark name + package, filtering to top-level benchmarks only"""
        by_benchmark = defaultdict(list)

        # Reverse commits so index 0 is oldest (left side of graph)
        commits_reversed = list(reversed(commits))

        for idx, commit in enumerate(commits_reversed):
            if not commit.success:
                continue
            for metric in commit.metrics:
                # Only include top-level benchmarks for cleaner graphs
                if self._is_top_level_benchmark(metric.name):
                    # Create unique key: package + benchmark name
                    key = f"{metric.package}::{metric.name}"
                    by_benchmark[key].append((idx, metric))

        return dict(by_benchmark)

    def generate(self, commits: List[CommitBenchmark]):
        """Generate all performance graphs"""
        by_benchmark = self._organize_by_benchmark(commits)

        if not by_benchmark:
            print("No benchmark data to visualize", file=sys.stderr)
            return

        # Generate metric graphs (all benchmarks combined)
        self._generate_metric_graph(by_benchmark, "iterations", "Iterations", "iterations.png")
        self._generate_metric_graph(by_benchmark, "ns_per_op", "Latency (ns/op)", "latency.png")
        self._generate_metric_graph(by_benchmark, "bytes_per_op", "Memory (B/op)", "memory.png")
        self._generate_metric_graph(by_benchmark, "allocs_per_op", "Allocations (allocs/op)", "allocs.png")

        print(f"\n✓ Graphs saved to {self.output_dir}/", file=sys.stderr)

    def _generate_metric_graph(self, by_benchmark: Dict, metric_name: str, ylabel: str, filename: str):
        """Generate a single metric graph with all benchmarks"""
        plt.figure(figsize=(16, 10))

        # Helper to extract category and label from package::name key
        def get_category_and_label(key: str) -> Tuple[str, str]:
            package, name = key.split("::", 1)

            # Determine category
            if "internal/meta/exif" in package:
                category = "EXIF"
                label = f"EXIF: {name}"
            elif "internal/meta/iptc" in package:
                category = "IPTC"
                label = f"IPTC: {name}"
            elif "internal/meta/xmp" in package:
                category = "XMP"
                label = f"XMP: {name}"
            elif "internal/meta/icc" in package:
                category = "ICC"
                label = f"ICC: {name}"
            elif "internal/format/jpeg" in package or "internal/container/jpeg" in package:
                category = "JPEG"
                label = f"JPEG: {name}"
            elif package.endswith("/imx") or package == "github.com/gomantics/imx":
                category = "API"
                # Simplify API labels
                if name.startswith("MetadataFrom"):
                    label = name.replace("MetadataFrom", "")
                elif name.startswith("Metadata_"):
                    label = name.replace("Metadata_", "")
                else:
                    label = name
            else:
                category = "Other"
                label = name

            return category, label

        # Group benchmarks by category for consistent colors
        categories = defaultdict(list)
        for key in by_benchmark.keys():
            category, _ = get_category_and_label(key)
            categories[category].append(key)

        color_map = {
            "API": "tab:blue",
            "EXIF": "tab:orange",
            "IPTC": "tab:green",
            "XMP": "tab:red",
            "ICC": "tab:purple",
            "JPEG": "tab:brown",
            "Other": "tab:gray",
        }

        # Plot each benchmark
        for category in sorted(categories.keys()):
            for key in sorted(categories[category]):
                data = by_benchmark.get(key, [])
                if not data:
                    continue

                indices = [idx for idx, _ in data]

                if metric_name == "iterations":
                    values = [m.iterations for _, m in data]
                elif metric_name == "ns_per_op":
                    values = [m.ns_per_op for _, m in data]
                elif metric_name == "bytes_per_op":
                    values = [m.bytes_per_op for _, m in data]
                elif metric_name == "allocs_per_op":
                    values = [m.allocs_per_op for _, m in data]
                else:
                    continue

                # Skip if all zeros
                if all(v == 0 for v in values):
                    continue

                _, label = get_category_and_label(key)
                plt.plot(indices, values, marker='o', label=label,
                        linewidth=2.5, markersize=6, color=color_map[category], alpha=0.8)

        # Format axes
        ax = plt.gca()

        # X-axis: commit indices
        ax.set_xlabel('Commit (oldest → newest)', fontsize=12, fontweight='bold')
        ax.set_xticks(range(len(set(idx for data in by_benchmark.values() for idx, _ in data))))

        # Y-axis: format based on metric type
        if metric_name == "ns_per_op":
            def format_latency(value, pos):
                if value >= 1_000_000_000:
                    return f'{value/1_000_000_000:.1f}s'
                elif value >= 1_000_000:
                    return f'{value/1_000_000:.1f}ms'
                elif value >= 1_000:
                    return f'{value/1_000:.1f}µs'
                else:
                    return f'{value:.0f}ns'
            ax.yaxis.set_major_formatter(plt.FuncFormatter(format_latency))
            ylabel = "Latency per Operation"
        elif metric_name == "bytes_per_op":
            def format_bytes(value, pos):
                if value >= 1024*1024:
                    return f'{value/(1024*1024):.1f}MB'
                elif value >= 1024:
                    return f'{value/1024:.1f}KB'
                else:
                    return f'{value:.0f}B'
            ax.yaxis.set_major_formatter(plt.FuncFormatter(format_bytes))
            ylabel = "Memory per Operation"
        elif metric_name == "iterations":
            def format_iterations(value, pos):
                if value >= 1_000_000:
                    return f'{value/1_000_000:.1f}M'
                elif value >= 1_000:
                    return f'{value/1_000:.1f}K'
                else:
                    return f'{value:.0f}'
            ax.yaxis.set_major_formatter(plt.FuncFormatter(format_iterations))
            ylabel = "Iterations"
        else:  # allocs_per_op
            ylabel = "Allocations per Operation"

        plt.ylabel(ylabel, fontsize=12, fontweight='bold')
        plt.title(f'Performance History - {ylabel}', fontsize=14, fontweight='bold', pad=20)
        plt.legend(loc='best', fontsize=9, framealpha=0.9, ncol=2)
        plt.grid(True, alpha=0.3, linestyle='--', linewidth=0.5)
        plt.tight_layout()
        plt.savefig(self.output_dir / filename, dpi=120, bbox_inches='tight')
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

    # Always run current benchmarks first with streaming output
    print("Running benchmarks on current commit...\n", file=sys.stderr)
    success, metrics, output = runner.run_current_benchmarks_streaming()

    if not success:
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
