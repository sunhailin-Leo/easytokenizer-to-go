# easytokenizer-to-go
Golang binding for https://github.com/zejunwang1/easytokenizer

**But in this project, we make compatible easytokenizer to 0.2.0, and rename some functions, because of C++ function override**

### Version

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