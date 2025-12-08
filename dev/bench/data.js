window.BENCHMARK_DATA = {
  "lastUpdate": 1765216958732,
  "repoUrl": "https://github.com/gomantics/imx",
  "entries": {
    "Benchmark": [
      {
        "commit": {
          "author": {
            "email": "puneetrai04@gmail.com",
            "name": "Puneet Rai",
            "username": "rpuneet"
          },
          "committer": {
            "email": "puneetrai04@gmail.com",
            "name": "Puneet Rai",
            "username": "rpuneet"
          },
          "distinct": true,
          "id": "fe0e773a857ab9b8cb5ebee7ca743c12feb3e1e6",
          "message": "fix(ci): simplify benchmark workflow with proper gh-pages setup\n\n- Created gh-pages branch manually (required first step)\n- Simplified workflow back to standard github-action-benchmark config\n- Removed peaceiris/actions-gh-pages (not needed, action handles it)\n- Set auto-push: true for automatic deployment\n\nNow that gh-pages branch exists, the action will work correctly.\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
          "timestamp": "2025-12-08T23:21:10+05:30",
          "tree_id": "b272f8595aa12a3fb1cada05cdaafd20b39c9dea",
          "url": "https://github.com/gomantics/imx/commit/fe0e773a857ab9b8cb5ebee7ca743c12feb3e1e6"
        },
        "date": 1765216313176,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkMetadataFromFile",
            "value": 662751,
            "unit": "ns/op\t  601734 B/op\t    3345 allocs/op",
            "extra": "1750 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - ns/op",
            "value": 662751,
            "unit": "ns/op",
            "extra": "1750 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - B/op",
            "value": 601734,
            "unit": "B/op",
            "extra": "1750 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - allocs/op",
            "value": 3345,
            "unit": "allocs/op",
            "extra": "1750 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse",
            "value": 5333,
            "unit": "ns/op\t   47920 B/op\t      24 allocs/op",
            "extra": "254768 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - ns/op",
            "value": 5333,
            "unit": "ns/op",
            "extra": "254768 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - B/op",
            "value": 47920,
            "unit": "B/op",
            "extra": "254768 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "254768 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse",
            "value": 2483,
            "unit": "ns/op\t    1632 B/op\t      31 allocs/op",
            "extra": "478708 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - ns/op",
            "value": 2483,
            "unit": "ns/op",
            "extra": "478708 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - B/op",
            "value": 1632,
            "unit": "B/op",
            "extra": "478708 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - allocs/op",
            "value": 31,
            "unit": "allocs/op",
            "extra": "478708 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse",
            "value": 7.189,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "167427402 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - ns/op",
            "value": 7.189,
            "unit": "ns/op",
            "extra": "167427402 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "167427402 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "167427402 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse",
            "value": 3876,
            "unit": "ns/op\t    4632 B/op\t      54 allocs/op",
            "extra": "298369 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - ns/op",
            "value": 3876,
            "unit": "ns/op",
            "extra": "298369 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - B/op",
            "value": 4632,
            "unit": "B/op",
            "extra": "298369 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - allocs/op",
            "value": 54,
            "unit": "allocs/op",
            "extra": "298369 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse",
            "value": 55965,
            "unit": "ns/op\t   34395 B/op\t     464 allocs/op",
            "extra": "21330 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - ns/op",
            "value": 55965,
            "unit": "ns/op",
            "extra": "21330 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - B/op",
            "value": 34395,
            "unit": "B/op",
            "extra": "21330 times\n4 procs"
          },
          {
            "name": "BenchmarkParser_Parse - allocs/op",
            "value": 464,
            "unit": "allocs/op",
            "extra": "21330 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "puneetrai04@gmail.com",
            "name": "Puneet Rai",
            "username": "rpuneet"
          },
          "committer": {
            "email": "puneetrai04@gmail.com",
            "name": "Puneet Rai",
            "username": "rpuneet"
          },
          "distinct": true,
          "id": "cb3bd53279b1affe41fa8f724aaabdb2ed48dd58",
          "message": "refactor(bench): reorganize benchmarks with clear naming and consolidate testdata\n\nChanges:\n- Rename benchmark functions to avoid confusion:\n  * BenchmarkParser_Parse -> BenchmarkJPEGParse (jpeg)\n  * BenchmarkParser_Parse -> BenchmarkEXIFParse (exif)\n  * BenchmarkParser_Parse -> BenchmarkIPTCParse (iptc)\n  * BenchmarkParser_Parse -> BenchmarkXMPParse (xmp)\n  * BenchmarkParser_Parse -> BenchmarkICCParse (icc)\n- Add comprehensive high-level API benchmarks:\n  * BenchmarkMetadataFromBytes\n  * BenchmarkMetadataFromReader\n  * BenchmarkMetadata_Tag\n  * BenchmarkMetadata_GetAll\n  * BenchmarkMetadata_Each\n- Consolidate testdata: move cmd/imx/testdata to root testdata\n- Remove benchmark history script (scripts/bench.py) - focus on continuous benchmarking\n- Update Makefile: remove bench-history target\n- Remove CODECOV_SETUP.md (not needed)\n- Update README with latest benchmark results\n- Update CONTRIBUTING.md: simplify benchmarking section\n\nBenchmark names are now clear and distinct, avoiding confusion in CI dashboard.\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
          "timestamp": "2025-12-08T23:31:45+05:30",
          "tree_id": "a2d4981d6f74c2bc99279db7ddc8f44039590614",
          "url": "https://github.com/gomantics/imx/commit/cb3bd53279b1affe41fa8f724aaabdb2ed48dd58"
        },
        "date": 1765216958421,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkMetadataFromFile",
            "value": 668609,
            "unit": "ns/op\t  601755 B/op\t    3345 allocs/op",
            "extra": "1717 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - ns/op",
            "value": 668609,
            "unit": "ns/op",
            "extra": "1717 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - B/op",
            "value": 601755,
            "unit": "B/op",
            "extra": "1717 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromFile - allocs/op",
            "value": 3345,
            "unit": "allocs/op",
            "extra": "1717 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromBytes",
            "value": 626960,
            "unit": "ns/op\t  601428 B/op\t    3342 allocs/op",
            "extra": "1909 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromBytes - ns/op",
            "value": 626960,
            "unit": "ns/op",
            "extra": "1909 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromBytes - B/op",
            "value": 601428,
            "unit": "B/op",
            "extra": "1909 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromBytes - allocs/op",
            "value": 3342,
            "unit": "allocs/op",
            "extra": "1909 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromReader",
            "value": 595587,
            "unit": "ns/op\t  601582 B/op\t    3342 allocs/op",
            "extra": "1930 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromReader - ns/op",
            "value": 595587,
            "unit": "ns/op",
            "extra": "1930 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromReader - B/op",
            "value": 601582,
            "unit": "B/op",
            "extra": "1930 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadataFromReader - allocs/op",
            "value": 3342,
            "unit": "allocs/op",
            "extra": "1930 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Tag",
            "value": 15.73,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "78228626 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Tag - ns/op",
            "value": 15.73,
            "unit": "ns/op",
            "extra": "78228626 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Tag - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "78228626 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Tag - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "78228626 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_GetAll",
            "value": 154.5,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "7930232 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_GetAll - ns/op",
            "value": 154.5,
            "unit": "ns/op",
            "extra": "7930232 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_GetAll - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "7930232 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_GetAll - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "7930232 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Each",
            "value": 3095,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "392114 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Each - ns/op",
            "value": 3095,
            "unit": "ns/op",
            "extra": "392114 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Each - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "392114 times\n4 procs"
          },
          {
            "name": "BenchmarkMetadata_Each - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "392114 times\n4 procs"
          },
          {
            "name": "BenchmarkJPEGParse",
            "value": 5135,
            "unit": "ns/op\t   47920 B/op\t      24 allocs/op",
            "extra": "269564 times\n4 procs"
          },
          {
            "name": "BenchmarkJPEGParse - ns/op",
            "value": 5135,
            "unit": "ns/op",
            "extra": "269564 times\n4 procs"
          },
          {
            "name": "BenchmarkJPEGParse - B/op",
            "value": 47920,
            "unit": "B/op",
            "extra": "269564 times\n4 procs"
          },
          {
            "name": "BenchmarkJPEGParse - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "269564 times\n4 procs"
          },
          {
            "name": "BenchmarkEXIFParse",
            "value": 2550,
            "unit": "ns/op\t    1632 B/op\t      31 allocs/op",
            "extra": "430251 times\n4 procs"
          },
          {
            "name": "BenchmarkEXIFParse - ns/op",
            "value": 2550,
            "unit": "ns/op",
            "extra": "430251 times\n4 procs"
          },
          {
            "name": "BenchmarkEXIFParse - B/op",
            "value": 1632,
            "unit": "B/op",
            "extra": "430251 times\n4 procs"
          },
          {
            "name": "BenchmarkEXIFParse - allocs/op",
            "value": 31,
            "unit": "allocs/op",
            "extra": "430251 times\n4 procs"
          },
          {
            "name": "BenchmarkICCParse",
            "value": 7.176,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "167107771 times\n4 procs"
          },
          {
            "name": "BenchmarkICCParse - ns/op",
            "value": 7.176,
            "unit": "ns/op",
            "extra": "167107771 times\n4 procs"
          },
          {
            "name": "BenchmarkICCParse - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "167107771 times\n4 procs"
          },
          {
            "name": "BenchmarkICCParse - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "167107771 times\n4 procs"
          },
          {
            "name": "BenchmarkIPTCParse",
            "value": 3921,
            "unit": "ns/op\t    4632 B/op\t      54 allocs/op",
            "extra": "305524 times\n4 procs"
          },
          {
            "name": "BenchmarkIPTCParse - ns/op",
            "value": 3921,
            "unit": "ns/op",
            "extra": "305524 times\n4 procs"
          },
          {
            "name": "BenchmarkIPTCParse - B/op",
            "value": 4632,
            "unit": "B/op",
            "extra": "305524 times\n4 procs"
          },
          {
            "name": "BenchmarkIPTCParse - allocs/op",
            "value": 54,
            "unit": "allocs/op",
            "extra": "305524 times\n4 procs"
          },
          {
            "name": "BenchmarkXMPParse",
            "value": 56454,
            "unit": "ns/op\t   34397 B/op\t     464 allocs/op",
            "extra": "21242 times\n4 procs"
          },
          {
            "name": "BenchmarkXMPParse - ns/op",
            "value": 56454,
            "unit": "ns/op",
            "extra": "21242 times\n4 procs"
          },
          {
            "name": "BenchmarkXMPParse - B/op",
            "value": 34397,
            "unit": "B/op",
            "extra": "21242 times\n4 procs"
          },
          {
            "name": "BenchmarkXMPParse - allocs/op",
            "value": 464,
            "unit": "allocs/op",
            "extra": "21242 times\n4 procs"
          }
        ]
      }
    ]
  }
}