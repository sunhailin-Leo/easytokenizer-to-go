# easytokenizer-to-go

简体中文 | [English](README.md)

BERT WordPiece 分词器的纯 Go 实现，也是
https://github.com/zejunwang1/easytokenizer 的 Go 移植版。

[![Docs](https://pkg.go.dev/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://pkg.go.dev/github.com/sunhailin-Leo/easytokenizer-to-go)
[![Report Card](https://goreportcard.com/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://goreportcard.com/report/github.com/sunhailin-Leo/easytokenizer-to-go)

[![Benchmark](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml)
[![Lint Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml)
[![Security Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml)
[![Test](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml)
[![Vulnerability Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml)

### 环境要求

* Go 1.27 及以上。本版本使用了标准库的 `simd` 包，它在 Go 1.27 中仍位于
  `GOEXPERIMENT=simd` 实验开关之后，并且语言版本要求为 `go 1.27`。

### 版本

* version 0.3.0
  * **破坏性发布。** 本模块现在是纯 Go。所有 C/C++ 源码、cgo 绑定、
    `build.sh` 和 `CMakeLists.txt` 均已删除；代码中不再有
    `import "C"`，`CGO_ENABLED=0 go build ./...` 可以直接通过。不需要安装
    动态库，也不需要再设置
    `CGO_CXXFLAGS`/`CGO_LDFLAGS`/`LD_LIBRARY_PATH`。
  * 使用 `GOEXPERIMENT=simd` 构建时，字节扫描原语会通过标准库 `simd` 包
    向量化；不加该开关时代码同样能编译，行为完全一致。
  * 对外 API 未做任何改动：`NewTokenizer`、`Close`、`Encode`、
    `EncodeWithIds`、`WordPieceTokenize` 的签名保持不变，分词结果与偏移量
    与 v0.2.1 完全一致。与旧版本的对比见
    [`BENCHMARKS.md`](BENCHMARKS.md)。
  * 在上述接口之上，本版本追加了一批纯新增的可选能力：可返回错误的构造器
    （支持 `go:embed`）、句对与批量编码、带显式截断/填充策略的结构化
    `Encoding` 结果、可配置的特殊 token，以及码点到字节偏移的换算。
    v0.1.0 的方法完全不受影响。
  * v0.1.x/0.2.x 的 C++ 实现已不在仓库中，如需查阅历史实现请切到
    git tag `v0.2.1`。

* version 0.2.1
  * 更新部分 free 函数。

* version 0.2.0
  * 修复 `EncodeWithIds`、`WordPieceTokenize` 的错误返回。
  * 增加 GitHub Workflows。
  * 更新测试代码。

* version 0.1.0
  * 完成 `initTokenizer`、`encode`、`encodeWithIds`、`wordPieceTokenize`。
  * 完成 Go 侧 `NewTokenizer`、`Close`、`Encode`、`EncodeWithIds`、
    `WordPieceTokenize`。

### 使用方式

```go
package main

import (
	"log"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

func main() {
	tk := tokenizer.NewTokenizer("./test/bert-chinese-vocab.txt", true)
	defer tk.Close()

	// 定长 id 序列，不足部分补 0。
	ids := tk.Encode("广东省深圳市南山区腾讯滨海大厦", 48)
	log.Println(ids)

	// input ids、token type ids、attention mask 以及码点级偏移量。
	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds("广东省深圳市南山区腾讯滨海大厦", 48)
	log.Println(inputIds, tokenTypeIds, attentionMask, offsets)

	// 原始 token，"##" 前缀表示这是一个续接片段。
	tokens, offsets := tk.WordPieceTokenize("广东省深圳市南山区腾讯滨海大厦")
	log.Println(tokens, offsets)
}
```

`NewTokenizer` 在词表文件读取失败时会 panic。分词器创建后是只读的，因此
一个实例可以被任意多个 goroutine 共享。`Close` 仅为兼容旧 API 而保留：
它不释放任何资源，重复调用安全，调用之后分词器依然可用。

#### 返回错误的构造器

```go
tk, err := tokenizer.NewTokenizerWithOptions("./vocab.txt", true)
tk, err := tokenizer.NewTokenizerFromReader(r, true)

//go:embed vocab.txt
var vocabFS embed.FS
tk, err := tokenizer.NewTokenizerFromFS(vocabFS, "vocab.txt", true)
```

`NewTokenizer` 是唯一会 panic 的构造器，它内部委托给
`NewTokenizerWithOptions`。

#### 偏移量的单位是码点

`EncodeWithIds` 与 `WordPieceTokenize` 返回码点偏移量。`tk.OffsetType()` 给出
单位，`tokenizer.ByteOffset` 把它换算成可用于切片的字节偏移：

```go
tokens, offsets := tk.WordPieceTokenize(text)
for i, token := range tokens {
	start := tokenizer.ByteOffset(text, offsets[2*i])
	end := tokenizer.ByteOffset(text, offsets[2*i+1])
	log.Println(token, text[start:end])
}
```

#### 句对与批量

```go
// [CLS] question [SEP] answer [SEP]，并补齐到 128。
ids := tk.EncodePair(question, answer, 128)

// token type id：question 为 0，answer 为 1。
ids, typeIds, mask, offsets := tk.EncodePairWithIds(question, answer, 128)

batch := tk.EncodeBatch([]string{"one", "two"}, 32)
```

#### 结构化结果与编码选项

`EncodeWithOptions`、`EncodePairWithOptions`、`EncodeBatchWithOptions` 返回
`Encoding`，其中包含 ids、token type ids、attention mask、token 字符串与
偏移量，并可显式指定截断和填充策略：

```go
enc := tk.EncodeWithOptions(text, tokenizer.EncodingOptions{MaxSeqLength: 128})
enc = tk.EncodePairWithOptions(question, answer, tokenizer.EncodingOptions{
	MaxSeqLength: 128,
	Truncation:   tokenizer.TruncateOnlySecond, // 也可用 TruncateLongestFirst / TruncateOnlyFirst
	Padding:      tokenizer.PadNone,            // 不填充，长度即真实长度
})
```

`MaxSeqLength` 为 0 表示既不截断也不填充。与固定补 0 的旧方法不同，这组 API
使用分词器自己的 PAD id 进行填充。

#### 自定义特殊 token

```go
tk, err := tokenizer.NewTokenizerFromFS(fsys, "vocab.txt", true,
	tokenizer.WithSpecialTokens(tokenizer.SpecialTokens{
		Pad: "[PAD]", CLS: "<s>", SEP: "</s>", UNK: "<unk>", Mask: "",
	}))
```

字段为空表示该 token 不存在；`tokenizer.DefaultSpecialTokens()` 返回 BERT 的
默认写法。

### SIMD

`go build` 和 `go test` 默认走可移植的标量实现。要启用向量化内核：

```bash
GOEXPERIMENT=simd go build ./...
GOEXPERIMENT=simd go test ./...
```

向量宽度在运行时取自 `simd.VectorBitSize()`，因此同一份源码既能跑在
128 位 Neon 上，也能跑在 512 位 AVX-512 上。该构建是可选的，产出的 token
完全相同；内核加速表、它加速了什么、没加速什么，以及哪些原语被刻意保留为
标量实现，都见 [`BENCHMARKS.md`](BENCHMARKS.md)。

### 目录结构

| 路径                                                             | 内容                                    |
|----------------------------------------------------------------|---------------------------------------|
| `tokenizer.go`                                                 | 冻结的 v0.1.0 API                       |
| `options.go`, `encoding.go`                                    | 新增 API：构造器、选项、句对、批量、`Encoding`      |
| `basic.go`, `wordpiece.go`                                     | 分词器主体                                 |
| `trie.go`, `vocab.go`                                          | 词表及其字节 trie                           |
| `ascii.go`, `codepoint.go`, `normalize.go`                     | Unicode 辅助函数：ASCII 表、码点分类、NFD + Mn 剥离 |
| `scan_simd.go`, `scan_nosimd.go`                               | 扫描原语，按构建标签二选一                         |
| `test/`                                                        | 词表、黄金测试与基准测试                          |
