package main

import (
	"os"
	"testing"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

func BenchmarkChineseEncode(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.Encode(testChineseText, 48)
	}

	tk.Close()
}

func BenchmarkChineseEncodeWithIds(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.EncodeWithIds(testChineseText, 48)
	}

	tk.Close()
}

func BenchmarkChineseWordPieceTokenize(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.WordPieceTokenize(testChineseText)
	}

	tk.Close()
}

func BenchmarkThaiEncode(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testMultilingualVocabFileName, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.Encode(testThaiLanguage, 64)
	}

	tk.Close()
}

func BenchmarkThaiEncodeWithIds(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testMultilingualVocabFileName, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.EncodeWithIds(testThaiLanguage, 64)
	}

	tk.Close()
}

func BenchmarkThaiWordPieceTokenize(b *testing.B) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testMultilingualVocabFileName, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tk.WordPieceTokenize(testThaiLanguage)
	}

	tk.Close()
}
