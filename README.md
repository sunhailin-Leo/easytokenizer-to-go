# easytokenizer-to-go
Golang binding for https://github.com/zejunwang1/easytokenizer

**But in this project, we make compatible easytokenizer to 0.2.0, and rename some functions, because of C++ function override**

[![Docs](https://pkg.go.dev/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://pkg.go.dev/badge/github.com/sunhailin-Leo/easytokenizer-to-go)
[![Report Card](https://goreportcard.com/badge/github.com/sunhailin-Leo/easytokenizer-to-go)](https://goreportcard.com/badge/github.com/sunhailin-Leo/easytokenizer-to-go)

[![Benchmark](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/benchmark.yml)
[![Lint Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/lint.yml)
[![Security Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/sercurity.yml)
[![Test](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/test.yml)
[![Vulnerability Check](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/sunhailin-Leo/easytokenizer-to-go/actions/workflows/vulncheck.yml)



### Version

* version 0.2.0
  * Fix `EncodeWithIds`, `WordPieceTokenize` API return error result.
  * Add Github Workflows.
  * Update test code.

* version 0.1.0
  * Finish API `initTokenizer`, `encode`, `encodeWithIds`, `wordPieceTokenize` in CGO.
  * Finish API `NewTokenizer`, `Close`, `Encode`, `EncodeWithIds`, `WordPieceTokenize` in Golang.

### Build Library

* Linux/MacOS
	* `sh build.sh`

### Usage

* When building golang program, please add `export CGO_CXXFLAGS=-std=c++11` command before `go build / run / test ...`


### Example

* In example folder.