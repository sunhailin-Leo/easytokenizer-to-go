# Benchmarks

Comparison of the v0.2.1 cgo binding with the pure Go implementation that
replaced it in v0.3.0, with and without `GOEXPERIMENT=simd`. The pure Go
columns include the three optimization passes described under
[the v0.3.0 optimization pass](#the-v030-optimization-pass).

## How these numbers were produced

* Apple M3 Pro (12 cores), macOS, `go1.27.1 darwin/arm64`
* `go test -run '^$' -bench . -benchmem -count 6`, median of the six runs
* The v0.2.1 numbers were taken from a separate compiled binary linked against
  a `libtokenizer` built with `-O1`. The C++ sources crash with SIGBUS inside
  `initTokenizer` when built at `-O2`/`-O3`; that defect is why the baseline
  had to be built at `-O1`, and it is one of the reasons the C++ sources were
  dropped in 0.3.0. The 0.2.1 sources remain available at git tag `v0.2.1`.
* The English workload has no v0.2.1 column: the baseline binary predates the
  benchmark, and the v0.2.1 tokenizer could not be rebuilt with it
* The v0.2.1 column was measured in an earlier session. The pure Go and
  `GOEXPERIMENT=simd` columns were re-measured after the last optimization
  pass below, each as an interleaved A/B in two copies of the tree so that
  consecutive runs share the same host conditions. The host was under
  background load (load average ~5), which makes the pure Go numbers, if
  anything, pessimistic

Reproduce with:

```bash
cd test
go test -run '^$' -bench . -benchmem -count 6 ./...                      # pure Go
GOEXPERIMENT=simd go test -run '^$' -bench . -benchmem -count 6 ./...    # pure Go + SIMD
```

## End to end

`Go vs cgo` above 1.00x means the pure Go build is faster. `simd vs Go` above
1.00x means the vectorised build is faster.

| benchmark                | cgo 0.2.1 | pure Go  | pure Go + GOEXPERIMENT=simd | Go vs cgo | simd vs Go |
|--------------------------|-----------|----------|-----------------------------|-----------|------------|
| ChineseEncode            | 1,336 ns  | 729 ns   | 758 ns                      | 1.83x     | 0.962x     |
| ChineseEncodeWithIds     | 1,940 ns  | 811 ns   | 826 ns                      | 2.39x     | 0.982x     |
| ChineseWordPieceTokenize | 1,690 ns  | 734 ns   | 766 ns                      | 2.30x     | 0.958x     |
| EnglishEncode            | n/a       | 7,770 ns | 7,710 ns                    | n/a       | 1.008x     |
| EnglishEncodeWithIds     | n/a       | 8,005 ns | 8,004 ns                    | n/a       | 1.000x     |
| EnglishWordPieceTokenize | n/a       | 8,150 ns | 8,244 ns                    | n/a       | 0.989x     |
| ThaiEncode               | 4,686 ns  | 3,393 ns | 3,574 ns                    | 1.38x     | 0.949x     |
| ThaiEncodeWithIds        | 5,494 ns  | 3,483 ns | 3,553 ns                    | 1.58x     | 0.980x     |
| ThaiWordPieceTokenize    | 5,986 ns  | 3,594 ns | 3,578 ns                    | 1.67x     | 1.004x     |

The pure Go build wins on every benchmark now, most of all on Chinese, where
the FFI boundary and the C++ trie dominated. `ThaiEncode` was the last tie
(0.98x after the first pass); the passes below moved it to 1.38x.

The vectorised build lands within a few percent of the plain one in both
directions - 0.949x to 1.008x across the nine benchmarks, with the spread
between runs on this host larger than the difference between the builds. The
kernels it accelerates are a small share of total runtime, so
`GOEXPERIMENT=simd` is best understood as a build that is not slower, rather
than one that is dramatically faster; the kernel table below is where the
vector paths actually show up.

## Allocations

| benchmark                | cgo 0.2.1    | pure Go       | pure Go + SIMD |
|--------------------------|--------------|---------------|----------------|
| ChineseEncode            | 192 B / 1    | 1,320 B / 9   | 1,320 B / 9    |
| ChineseEncodeWithIds     | 752 B / 9    | 1,704 B / 11  | 1,704 B / 11   |
| ChineseWordPieceTokenize | 448 B / 20   | 1,304 B / 8   | 1,304 B / 8    |
| EnglishEncode            | n/a          | 9,096 B / 19  | 9,096 B / 19   |
| EnglishEncodeWithIds     | n/a          | 10,120 B / 21 | 10,120 B / 21  |
| EnglishWordPieceTokenize | n/a          | 11,592 B / 18 | 11,592 B / 18  |
| ThaiEncode               | 256 B / 1    | 3,144 B / 16  | 3,144 B / 16   |
| ThaiEncodeWithIds        | 1,104 B / 9  | 3,656 B / 18  | 3,656 B / 18   |
| ThaiWordPieceTokenize    | 1,072 B / 39 | 4,184 B / 15  | 4,184 B / 15   |

The 0.2.1 binding allocated its result buffers in C++ and handed them back as one
Go slice, so a single Go allocation could carry kilobytes and the comparison is
not like for like: cgo allocated fewer, larger blocks. The pure Go build
allocates every intermediate slice on the Go heap, and after the passes below it
allocates 7 to 12 fewer times per call than the first v0.3.0 build did, and 4
fewer than the build that preceded the last pass.

## The v0.3.0 optimization pass

The first pure Go build was measured, profiled, and then optimised in three
passes: R1-R4 with P1, then P2, then P4 with P5. Each pass was measured against
the build that preceded it, so the tables below are not affected by the drift
between sessions.

First pass, both columns measured back to back in the same session:

| benchmark                |    before | after R1-R4+P1 | change | allocs before | allocs after |
|--------------------------|----------:|---------------:|-------:|--------------:|-------------:|
| ChineseEncode            |  1,081 ns |         962 ns | -11.0% |            16 |            9 |
| ChineseEncodeWithIds     |  1,133 ns |       1,024 ns |  -9.6% |            18 |           11 |
| ChineseWordPieceTokenize |    897 ns |         798 ns | -11.0% |            14 |            8 |
| EnglishEncode            | 10,788 ns |      10,305 ns |  -4.5% |            31 |           19 |
| EnglishEncodeWithIds     | 11,373 ns |      10,530 ns |  -7.4% |            33 |           21 |
| EnglishWordPieceTokenize |  9,380 ns |       9,080 ns |  -3.2% |            29 |           18 |
| ThaiEncode               |  5,142 ns |       4,872 ns |  -5.2% |            29 |           20 |
| ThaiEncodeWithIds        |  5,266 ns |       4,924 ns |  -6.5% |            31 |           22 |
| ThaiWordPieceTokenize    |  4,799 ns |       4,603 ns |  -4.1% |            27 |           19 |

What changed:

* `Encode` and `EncodeWithIds` fill their output slice in place instead of
  building an intermediate id slice and copying it, and the WordPiece result
  slices are sized from the base token count.
* `byteTrie.child` scans the edges of narrow nodes linearly instead of by
  binary search. 91% of the nodes of the multilingual vocabulary have three or
  fewer children.
* `buildTrie` sizes its node and edge slices exactly, from the sorted key list,
  instead of reserving four per key.
* `loadVocab` parses the vocabulary buffer directly instead of materialising a
  `[]string` of lines first, and sizes the id map and token slice from the line
  count.

Second pass, an interleaved A/B in two copies of the tree that differ only in
this change, median of six runs each:

| benchmark                |  R1-R4+P1 |     + P2 |       change |
|--------------------------|----------:|---------:|-------------:|
| EnglishWordPieceTokenize |  9,533 ns | 8,431 ns |       -11.6% |
| EnglishEncodeWithIds     | 11,029 ns | 9,787 ns |       -11.3% |
| EnglishEncode            | 10,426 ns | 9,550 ns |        -8.4% |
| ChineseWordPieceTokenize |    861 ns |   813 ns |        -5.6% |
| ThaiEncode               |  5,059 ns | 4,778 ns |        -5.6% |
| ThaiEncodeWithIds        |  5,030 ns | 4,836 ns |        -3.9% |
| ThaiWordPieceTokenize    |  4,549 ns | 4,382 ns |        -3.7% |
| ChineseEncodeWithIds     |  1,069 ns | 1,055 ns |        -1.3% |
| ChineseEncode            |    989 ns |   994 ns | +0.5%, noise |

P2 replaces the `"##" + piece` concatenation in the WordPiece loop with a trie
entry point that walks the `"##"` bytes and the suffix separately
(`longestPrefixParts`). The concatenation does not allocate - the compiler
stack-allocates it, 0 B/op in a probe - but it copies the whole remaining word
on every continuation attempt, which is why the win scales with how often
English and Thai words are split.

Third pass, an interleaved A/B in two copies of the tree that differ only in
this change, median of six runs each:

| benchmark                | R1-R4+P1+P2 |    + P4+P5 | change | allocs before | allocs after |
|--------------------------|------------:|-----------:|-------:|--------------:|-------------:|
| ThaiEncode               |    4,643 ns |   3,393 ns | -26.9% |            20 |           16 |
| ThaiEncodeWithIds        |    4,724 ns |   3,483 ns | -26.3% |            22 |           18 |
| ChineseEncode            |      963 ns |     729 ns | -24.3% |             9 |            9 |
| ChineseEncodeWithIds     |    1,038 ns |     811 ns | -21.9% |            11 |           11 |
| EnglishEncode            |    9,464 ns |   7,770 ns | -17.9% |            19 |           19 |
| EnglishEncodeWithIds     |    9,720 ns |   8,005 ns | -17.6% |            21 |           21 |
| ThaiWordPieceTokenize    |    4,325 ns |   3,594 ns | -16.9% |            19 |           15 |
| ChineseWordPieceTokenize |      798 ns |     734 ns |  -8.0% |             8 |            8 |
| EnglishWordPieceTokenize |    8,353 ns |   8,150 ns |  -2.4% |            18 |           18 |

The same A/B under `GOEXPERIMENT=simd` moved by -27.6% to -4.1%, so the pass is
not an artifact of the plain build.

What changed:

* The WordPiece walk now resolves the id of every token while it walks, so no
  token is looked up in the vocabulary twice: `Encode` and `EncodeWithIds` ask
  the walk for the ids and copy them into the result, instead of mapping every
  token through `vocab.id` afterwards, and the split loop carries the id the
  trie already returned. A caller that wants the strings (`WordPieceTokenize`)
  gets those from the same walk and never materialises the ids.
* The vocabulary is asked before the special token trie, so a word that is part
  of it skips a trie walk that could only ever fail on its first byte for
  anything not starting with `[`.
* `buildPosMap`, the codepoint to offset map of the split path, is skipped
  entirely for the words it would map to the identity. `loadVocab` records that
  property once per token (`posMapIdentity`, one byte per token) by walking the
  token the way `buildPosMap` does; the walk the split path already performs
  then answers it for the whole word, because the pieces tile it. On the Thai
  workload this removes 10.5% cumulative profile time and, with it, four
  allocations per call.

Construction cost and retained heap for the last pass, measured with the same
driver and the same interleaved A/B:

| vocabulary                                       | construct before | construct after | retained before | retained after |
|--------------------------------------------------|-----------------:|----------------:|----------------:|---------------:|
| `bert-multilingual-vocab.txt`, `doLowerCase=false` |         34.5 ms |         39.5 ms |       15.19 MiB |      15.31 MiB |
| `bert-multilingual-vocab.txt`, `doLowerCase=true`  |         34.5 ms |         51.5 ms |       15.19 MiB |      15.31 MiB |
| `bert-chinese-vocab.txt`, `doLowerCase=true`       |          3.6 ms |          4.4 ms |        1.78 MiB |       1.80 MiB |

The identity table costs one byte per token and the scan that fills it costs
5 ms on the 119,547 token vocabulary, or 17 ms when the tokenizer lowercases,
because only then does every codepoint have to be checked against its lowercase
NFD form. It is a one-time cost against a per-call win of 8-27%, and it is paid
once per tokenizer instead of once per token on every call.

Construction cost and retained heap, measured with a standalone driver
(`runtime.GC()` and `ReadMemStats` around `NewTokenizer`):

| vocabulary                                     | construct before | construct after | retained before | retained after |
|------------------------------------------------|-----------------:|----------------:|----------------:|---------------:|
| `bert-multilingual-vocab.txt` (119,547 tokens) |          41.0 ms |         35.9 ms |       18.64 MiB |      15.19 MiB |
| `bert-chinese-vocab.txt` (21,128 tokens)       |           3.9 ms |          3.6 ms |        3.76 MiB |       1.78 MiB |

The package retains no per-call state: 20,000 `Encode` calls move `HeapAlloc`
by less than 15 KB in either direction. What was measured and then rejected is
recorded under [Measured and rejected](#measured-and-rejected).

## Scanning kernels

The SIMD build vectorises `alnumRun` and `countCodepoints`; the microbenchmarks
below use inputs long enough to reach the vector path.

| kernel                   | input                   | pure Go  | GOEXPERIMENT=simd | ratio |
|--------------------------|-------------------------|----------|-------------------|-------|
| ScanAlnumRun             | 121 B, alnum run of 120 | 163.8 ns | 97.8 ns           | 1.67x |
| ScanAlnumRunShort        | 13 B                    | 17.7 ns  | 18.0 ns           | 0.98x |
| ScanCountCodepointsASCII | 240 B ASCII             | 83.6 ns  | 70.7 ns           | 1.18x |
| ScanCountCodepointsCJK   | 72 B CJK                | 45.2 ns  | 39.9 ns           | 1.13x |
| ScanCountCodepointsShort | 7 B                     | 3.5 ns   | 4.2 ns            | 0.84x |

Run them with `go test -run '^$' -bench Scan -benchmem .` from the repository
root.

### The search for the next special token

`nextSpecialStart`, the search for the `[` that starts a special token, is the
one primitive that is not a `simd`-package kernel. It calls `strings.IndexByte`,
whose arm64 implementation is hand-written NEON assembly that reduces its
comparison mask to a byte index in hardware. The scalar column below is
`BenchmarkScanNextSpecialStart*Scalar`, kept in `scan_bench_test.go` as the
record of what the stdlib call replaced:

| kernel                    | input                 | scalar loop | `strings.IndexByte` | ratio |
|---------------------------|-----------------------|-------------|---------------------|-------|
| ScanNextSpecialStart      | 246 B, no `[`         | 81.8 ns     | 4.31 ns             | 19.0x |
| ScanNextSpecialStartHit   | 246 B, `[` at the end | 80.0 ns     | 4.03 ns             | 19.8x |
| ScanNextSpecialStartShort | 7 B, no `[`           | 3.69 ns     | 2.04 ns             | 1.8x  |

The `strings.IndexByte` and scalar columns are measured in both builds and
agree within noise. An earlier session measured the 7 byte scalar reference
about twice as slow under `GOEXPERIMENT=simd` (7.4 ns) and saw it reproducibly;
it did not reproduce in the re-measurement above, so it is recorded as
unexplained rather than as a property of the build. Neither path is a
`simd`-package kernel, so the choice does not depend on the build.

### What is deliberately not vectorised

One primitive stays scalar in both builds, because measurement said the vector
version was not worth shipping:

* `allASCIIAlnum` scored 2.35x faster on a long all alphanumeric word but 20x
  slower on a word that fails on its sixth byte: a vector kernel that ORs its
  masks and inspects them once cannot exit early. Single words fail early far
  more often than not, and it is called on short inputs (single words, single
  vocabulary pieces) where any vector setup loses anyway.

`nextSpecialStart` was in this list too: a vector version of it scored 1.7x
**slower** than the plain loop, because turning a lane mask back into an index
needs `Mask8x16.ToBits`, which only `simd/archsimd` offers on amd64; arm64 has
no equivalent, so the vector version degenerates to storing the mask and
scanning all 16 lanes, which is exactly the work the scalar loop does. It no
longer needs the note: it delegates to `strings.IndexByte` now.

`countCodepoints` keeps its vector path despite the slowdown on 7 byte inputs,
because it always scans the whole input, the vector form wins on anything a
vector wide, and the 7 byte case is the scalar path plus one predictable
branch.

## Measured and rejected

The optimization passes above are done, and so is the usability work that
shipped with them: error returning constructors with `go:embed` support,
published offset units with a codepoint to byte conversion, sentence pairs,
batches, a structured `Encoding` result with explicit truncation and padding
strategies, configurable special tokens, and `Close` implemented and documented
as the compatibility no-op it always was. `Encode`, `EncodeWithIds` and
`WordPieceTokenize` kept their signatures and their output throughout, and
every constructor option resolves to an immutable value, because a live
tokenizer is safe for concurrent use and is race-tested.

What was measured and then dropped, and what was rejected without shipping, is
recorded here so that the same ground is not re-explored.

Dropped after measuring:

* **Caching vocabulary codepoint lengths.** `getCodepointNumber` rescans the
  selected vocabulary string on every split, but it stays below 0.6% cumulative
  in all three post-optimization profiles, and a cache would add about 0.5 MiB
  of retained memory to the multilingual vocabulary.

Rejected, resources:

| idea | why not |
|---|---|
| pool the returned `[]string`/`[]int32` | the API has no release step, so the package cannot know when the caller is done |
| share one backing array across the four `EncodeWithIds` results | caller-visible aliasing; independent slices are part of the contract |
| replace `map[string]int32` with a sorted slice + binary search | saves an estimated 3-6 MiB but turns hot lookups from ~O(1) into `O(log N)` string comparisons; a byte arena with a hash index is the better shape but is a large rewrite |
| shrink `trieNode` from 12 bytes | a packed `uint32 first, uint32 count, int32 value` still occupies 12 bytes; at most 0.3-1.2 MiB with a split value array, for extra indirection |
| an arena or `unique` interning for vocabulary strings | lifecycle complexity for a construction-time-only win |

Rejected, performance:

* Rewriting the basic-tokenizer state machine: 2-3% self time, and it is where
  the token boundaries are decided.
* Replacing `longestPrefix`/`prefixes` with a different splitting loop: every
  candidate piece still has to be validated.
* Anything that changes token or offset output: the golden tests exist to pin
  that down, and v0.3.0 shipped on the promise that output is unchanged.

The dependency surface is settled as well: `golang.org/x/text` stays, because
Go 1.27 adds no Unicode normalization API, and `unicode/utf8` and
`strings.ToLower` cannot replace NFD decomposition plus Mn stripping without
changing tokenizer output. Removing the dependency would mean shipping
decomposition tables.
