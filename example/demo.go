package main

import (
	"fmt"
	"os"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

func main() {
	testStr := "广东省深圳市南山区腾讯滨海大厦"
	testMaxSeqLen := 48

	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/example/bert-chinese-vocab.txt", true)
	fmt.Println(tk.Encode(testStr, testMaxSeqLen))
	tk.EncodeWithIds(testStr, testMaxSeqLen)
	tk.WordPieceTokenize(testStr)
	tk.Close()
}
