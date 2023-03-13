package main

import (
	"os"
	"testing"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

// BenchmarkEncode-12    	  385312	      2620 ns/op	     208 B/op	       2 allocs/op
func BenchmarkEncode(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/bert-chinese-vocab.txt", true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.Encode("广东省深圳市南山区腾讯滨海大厦", 48)
	}
}

// BenchmarkEncodeWithIds-12    	  273579	      3726 ns/op	     128 B/op	      13 allocs/op
func BenchmarkEncodeWithIds(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/bert-chinese-vocab.txt", true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.EncodeWithIds("广东省深圳市南山区腾讯滨海大厦", 48)
	}
}

// BenchmarkWordPieceTokenize-12    	  283342	      4340 ns/op	     360 B/op	      23 allocs/op
func BenchmarkWordPieceTokenize(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/bert-chinese-vocab.txt", true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.WordPieceTokenize("广东省深圳市南山区腾讯滨海大厦")
	}
}
