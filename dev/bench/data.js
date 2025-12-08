window.BENCHMARK_DATA = {
  "lastUpdate": 1765216313672,
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
      }
    ]
  }
}