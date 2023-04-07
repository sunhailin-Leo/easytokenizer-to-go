package main

import (
	"log"
	"os"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

const (
	testChineseVocabFilename      string = "bert-chinese-vocab.txt"
	testMultilingualVocabFileName string = "bert-multilingual-vocab.txt"

	testChineseText  string = "广东省深圳市南山区腾讯滨海大厦"
	testThaiLanguage string = "นครปฐม เมืองนครปฐม ถนนขาด เลขที่ 69 หมู่ 1 ซ. - - ถ. -"
)

func testChineseTokenizer() {
	testStr := "广东省深圳市南山区腾讯滨海大厦"
	testMaxSeqLen := 48
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/test/bert-chinese-vocab.txt", true)

	// Encode
	encodeRes := tk.Encode(testStr, testMaxSeqLen)
	log.Println("Encode: ", encodeRes, len(encodeRes), len(encodeRes) == testMaxSeqLen)

	// EncodeWithIds
	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds(testStr, testMaxSeqLen)
	log.Println("InputIds: ", inputIds)
	log.Println("TokenTypeIds: ", tokenTypeIds)
	log.Println("AttentionMas: ", attentionMask)
	log.Println("Offsets: ", offsets)

	// WordPieceTokenize
	tokens, wordPieceOffsets := tk.WordPieceTokenize(testStr)
	log.Println("Tokens: ", tokens)
	log.Println("Offsets: ", wordPieceOffsets)

	// Release
	tk.Close()
}

func testMultilingualTokenizer() {
	testStr := "นครปฐม เมืองนครปฐม ถนนขาด เลขที่ 69 หมู่ 1 ซ. - - ถ. -"
	testMaxSeqLen := 64
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/test/bert-multilingual-vocab.txt", false)

	// Encode
	encodeRes := tk.Encode(testStr, testMaxSeqLen)
	log.Println("Encode: ", encodeRes, len(encodeRes), len(encodeRes) == testMaxSeqLen)

	// EncodeWithIds
	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds(testStr, testMaxSeqLen)
	log.Println("InputIds: ", inputIds)
	log.Println("TokenTypeIds: ", tokenTypeIds)
	log.Println("AttentionMas: ", attentionMask)
	log.Println("Offsets: ", offsets)

	// WordPieceTokenize
	tokens, offsets := tk.WordPieceTokenize(testStr)
	log.Println("Tokens: ", tokens)
	log.Println("Offsets: ", offsets)

	// Release
	tk.Close()
}

func main() {
	// Chinese
	testChineseTokenizer()

	log.Println("#################################################################################################")

	// Other Language
	testMultilingualTokenizer()
}
