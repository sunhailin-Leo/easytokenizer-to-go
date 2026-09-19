# easytokenizer-to-go

English | [简体中文](README.zh-CN.md)

Pure Go implementation of the BERT WordPiece tokenizer, and the Go port of
https://github.com/zejunwang1/easytokenizer

[![Docs](https://pkg.go.dev/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://pkg.go.dev/github.com/sunhailin-Leo/easytokenizer-to-go)
[![Report Card](https://goreportcard.com/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://goreportcard.com/report/github.com/sunhailin-Leo/easytokenizer-to-go)

[![Benchmark](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml)
[![Lint Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml)
[![Security Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml)
[![Test](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml)
[![Vulnerability Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml)

### Requirements

* Go 1.27 or later. This release uses the standard library `simd` package, which
  only exists behind `GOEXPERIMENT=simd` in Go 1.27, and takes `go 1.27` as its
  language version.

### Version

* version 0.3.0
  * **Breaking release.** The module is pure Go. Every C and C++ source, the
    cgo binding, `build.sh` and `CMakeLists.txt` have been deleted; there is no
    `import "C"` anywhere, and `CGO_ENABLED=0 go build ./...` succeeds. There
    is no shared library to install and no
    `CGO_CXXFLAGS`/`CGO_LDFLAGS`/`LD_LIBRARY_PATH` to set.
  * Building with `GOEXPERIMENT=simd` vectorises the byte scanning primitives
    with the standard library `simd` package. The code compiles and behaves
    identically without it.
  * The exported API is unchanged: `NewTokenizer`, `Close`, `Encode`,
    `EncodeWithIds` and `WordPieceTokenize` keep their signatures, and the
    tokenizer produces the same tokens and offsets as v0.2.1. See
    [`BENCHMARKS.md`](BENCHMARKS.md) for the comparison against the old build.
  * On top of that interface the release adds optional, purely additive
    surface: error returning constructors with `go:embed` support, sentence
    pairs and batches, a structured `Encoding` result with explicit truncation
    and padding strategies, configurable special tokens, and codepoint to byte
    offset conversion. The v0.1.0 methods are untouched by all of it.
  * The v0.1.x/0.2.x C++ implementation is no longer in the tree. It lives at
    git tag `v0.2.1` for anyone who needs the historical reference.

* version 0.2.1
  * Update some free function. 

* version 0.2.0
  * Fix `EncodeWithIds`, `WordPieceTokenize` API return error result.
  * Add Github Workflows.
  * Update test code.

* version 0.1.0
  * Finish API `initTokenizer`, `encode`, `encodeWithIds`, `wordPieceTokenize`.
  * Finish API `NewTokenizer`, `Close`, `Encode`, `EncodeWithIds`, `WordPieceTokenize` in Golang.

### Usage

```go
package main

import (
	"log"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

func main() {
	tk := tokenizer.NewTokenizer("./test/bert-chinese-vocab.txt", true)
	defer tk.Close()

	// Fixed size id sequence, padded with zeros.
	ids := tk.Encode("广东省深圳市南山区腾讯滨海大厦", 48)
	log.Println(ids)

	// Ids, token type ids, attention mask and codepoint level offsets.
	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds("广东省深圳市南山区腾讯滨海大厦", 48)
	log.Println(inputIds, tokenTypeIds, attentionMask, offsets)

	// Raw tokens, with "##" marking a continuation piece.
	tokens, offsets := tk.WordPieceTokenize("广东省深圳市南山区腾讯滨海大厦")
	log.Println(tokens, offsets)
}
```

`NewTokenizer` panics if the vocabulary file cannot be read. A tokenizer is
read-only once created, so one instance may be shared by any number of
goroutines. `Close` is a no-op kept for API compatibility: it releases nothing,
is safe to call more than once, and a tokenizer stays usable after it.

#### Constructors that return an error

```go
tk, err := tokenizer.NewTokenizerWithOptions("./vocab.txt", true)
tk, err := tokenizer.NewTokenizerFromReader(r, true)

//go:embed vocab.txt
var vocabFS embed.FS
tk, err := tokenizer.NewTokenizerFromFS(vocabFS, "vocab.txt", true)
```

`NewTokenizer` is the only constructor that panics; it delegates to
`NewTokenizerWithOptions`.

#### Offsets are codepoints

`EncodeWithIds` and `WordPieceTokenize` report codepoint offsets.
`tk.OffsetType()` returns the unit, and `tokenizer.ByteOffset` converts one into
the byte offset to slice the input with:

```go
tokens, offsets := tk.WordPieceTokenize(text)
for i, token := range tokens {
	start := tokenizer.ByteOffset(text, offsets[2*i])
	end := tokenizer.ByteOffset(text, offsets[2*i+1])
	log.Println(token, text[start:end])
}
```

#### Sentence pairs and batches

```go
// [CLS] question [SEP] answer [SEP], padded to 128.
ids := tk.EncodePair(question, answer, 128)

// Token type ids are 0 for the question and 1 for the answer.
ids, typeIds, mask, offsets := tk.EncodePairWithIds(question, answer, 128)

batch := tk.EncodeBatch([]string{"one", "two"}, 32)
```

#### Structured results and encoding options

`EncodeWithOptions`, `EncodePairWithOptions` and `EncodeBatchWithOptions` return
an `Encoding` holding the ids, token type ids, attention mask, token strings and
offsets, and take the truncation and padding strategy explicitly:

```go
enc := tk.EncodeWithOptions(text, tokenizer.EncodingOptions{MaxSeqLength: 128})
enc = tk.EncodePairWithOptions(question, answer, tokenizer.EncodingOptions{
	MaxSeqLength: 128,
	Truncation:   tokenizer.TruncateOnlySecond, // or TruncateLongestFirst / TruncateOnlyFirst
	Padding:      tokenizer.PadNone,            // exact length, no padding
})
```

A `MaxSeqLength` of zero means no limit and no padding. Unlike the frozen
methods, which fill with zero, the options API pads with the PAD id of the
tokenizer.

#### Custom special tokens

```go
tk, err := tokenizer.NewTokenizerFromFS(fsys, "vocab.txt", true,
	tokenizer.WithSpecialTokens(tokenizer.SpecialTokens{
		Pad: "[PAD]", CLS: "<s>", SEP: "</s>", UNK: "<unk>", Mask: "",
	}))
```

An empty field removes that token, and
`tokenizer.DefaultSpecialTokens()` returns the BERT names.

### SIMD

`go build` and `go test` use a portable scalar implementation by default. To
build with the vectorised kernels:

```bash
GOEXPERIMENT=simd go build ./...
GOEXPERIMENT=simd go test ./...
```

The vector width is taken from `simd.VectorBitSize()` at run time, so the same
source runs on 128-bit Neon and on 512-bit AVX-512 machines. The build is
optional and produces identical tokens; see [`BENCHMARKS.md`](BENCHMARKS.md) for
the kernel table, for what it does and does not speed up, and for the
primitives that stay scalar on purpose.

### Layout

| path                                                           | contents                                                                |
|----------------------------------------------------------------|-------------------------------------------------------------------------|
| `tokenizer.go`                                                 | the frozen v0.1.0 API                                                   |
| `options.go`, `encoding.go`                                    | the additive API: constructors, options, pairs, batches, `Encoding`      |
| `basic.go`, `wordpiece.go`                                     | the tokenizer itself                                                    |
| `trie.go`, `vocab.go`                                          | the vocabulary and its byte trie                                        |
| `ascii.go`, `codepoint.go`, `normalize.go`                     | Unicode helpers: ASCII tables, codepoint categories, NFD + Mn stripping |
| `scan_simd.go`, `scan_nosimd.go`                               | the scanning primitives, per build                                      |
| `test/`                                                        | the vocabularies, golden tests and the benchmarks                       |
