window.BENCHMARK_DATA = {
  "lastUpdate": 1700536606545,
  "repoUrl": "https://github.com/sunhailin-Leo/easytokenizer-to-go",
  "entries": {
    "Benchmark": [
      {
        "commit": {
          "author": {
            "email": "379978424@qq.com",
            "name": "LeoSun",
            "username": "sunhailin-Leo"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e7b6f8c4b10548d1bbacb6113819f8c78ff034a3",
          "message": "Update benchmark.yml",
          "timestamp": "2023-11-21T11:16:01+08:00",
          "tree_id": "dfe5bf7f69e68a732291f11b50599dcacda2244d",
          "url": "https://github.com/sunhailin-Leo/easytokenizer-to-go/commit/e7b6f8c4b10548d1bbacb6113819f8c78ff034a3"
        },
        "date": 1700536606102,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkChineseEncode",
            "value": 1565,
            "unit": "ns/op\t     192 B/op\t       1 allocs/op",
            "extra": "772414 times\n4 procs"
          },
          {
            "name": "BenchmarkChineseEncodeWithIds",
            "value": 2330,
            "unit": "ns/op\t     752 B/op\t      12 allocs/op",
            "extra": "466432 times\n4 procs"
          },
          {
            "name": "BenchmarkChineseWordPieceTokenize",
            "value": 2887,
            "unit": "ns/op\t     440 B/op\t      21 allocs/op",
            "extra": "411492 times\n4 procs"
          },
          {
            "name": "BenchmarkThaiEncode",
            "value": 7475,
            "unit": "ns/op\t     256 B/op\t       1 allocs/op",
            "extra": "153140 times\n4 procs"
          },
          {
            "name": "BenchmarkThaiEncodeWithIds",
            "value": 8436,
            "unit": "ns/op\t    1104 B/op\t      12 allocs/op",
            "extra": "141429 times\n4 procs"
          },
          {
            "name": "BenchmarkThaiWordPieceTokenize",
            "value": 10459,
            "unit": "ns/op\t    1056 B/op\t      40 allocs/op",
            "extra": "114783 times\n4 procs"
          }
        ]
      }
    ]
  }
}