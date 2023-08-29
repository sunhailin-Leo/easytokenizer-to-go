package main

import (
	"os"
	"reflect"
	"testing"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

func TestChineseTokenizer(t *testing.T) {
	pwd, _ := os.Getwd()
	testMaxSeqLen := 48
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)

	encodeRes := tk.Encode(testChineseText, testMaxSeqLen)
	expectedEncodeRes := []int32{
		101, 2408, 691, 4689, 3918, 1766, 2356, 1298, 2255, 1277, 5596, 6380, 4012, 3862, 1920, 1336, 102,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(encodeRes, expectedEncodeRes) {
		t.Errorf("Expected '%v', but got '%v'", encodeRes, expectedEncodeRes)
	}

	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds(testChineseText, testMaxSeqLen)

	expectedInputIds := []int32{
		101, 2408, 691, 4689, 3918, 1766, 2356, 1298, 2255, 1277, 5596, 6380, 4012, 3862, 1920, 1336, 102,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(inputIds, expectedInputIds) {
		t.Errorf("Expected '%v', but got '%v'", inputIds, expectedInputIds)
	}

	expectedTokenTypeIds := []int32{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(tokenTypeIds, expectedTokenTypeIds) {
		t.Errorf("Expected '%v', but got '%v'", tokenTypeIds, expectedTokenTypeIds)
	}

	expectedAttentionMask := []int32{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(attentionMask, expectedAttentionMask) {
		t.Errorf("Expected '%v', but got '%v'", attentionMask, expectedAttentionMask)
	}

	expectedOffsets := []int32{
		0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11,
		12, 12, 13, 13, 14, 14, 15,
	}
	if !reflect.DeepEqual(offsets, expectedOffsets) {
		t.Errorf("Expected '%v', but got '%v'", offsets, expectedOffsets)
	}

	tokens, wordPieceOffsets := tk.WordPieceTokenize(testChineseText)

	expectedTokens := []string{"广", "东", "省", "深", "圳", "市", "南", "山", "区", "腾", "讯", "滨", "海", "大", "厦"}
	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Errorf("Expected '%v', but got '%v'", tokens, expectedTokens)
	}

	expectedWordPieceOffsets := []int32{
		0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11,
		12, 12, 13, 13, 14, 14, 15,
	}
	if !reflect.DeepEqual(wordPieceOffsets, expectedWordPieceOffsets) {
		t.Errorf("Expected '%v', but got '%v'", wordPieceOffsets, expectedWordPieceOffsets)
	}

	tk.Close()
}

func TestMultilingualTokenizer_Thai(t *testing.T) {
	pwd, _ := os.Getwd()
	testMaxSeqLen := 64
	tk := tokenizer.NewTokenizer(pwd+"/"+testMultilingualVocabFileName, false)

	encodeRes := tk.Encode(testThaiLanguage, testMaxSeqLen)
	expectedEncodeRes := []int32{
		101, 1417, 51752, 49292, 99671, 17405, 1450, 105710, 16000, 51752, 49292, 99671, 17405, 1414, 16000,
		16000, 80814, 17344, 22123, 1450, 20507, 80814, 18203, 12573, 1434, 17405, 86063, 122, 1403, 119, 118,
		118, 1414, 119, 118, 102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(encodeRes, expectedEncodeRes) {
		t.Errorf("Expected '%v', but got '%v'", encodeRes, expectedEncodeRes)
	}

	inputIds, tokenTypeIds, attentionMask, offsets := tk.EncodeWithIds(testThaiLanguage, testMaxSeqLen)

	expectedInputIds := []int32{
		101, 1417, 51752, 49292, 99671, 17405, 1450, 105710, 16000, 51752, 49292, 99671, 17405, 1414, 16000,
		16000, 80814, 17344, 22123, 1450, 20507, 80814, 18203, 12573, 1434, 17405, 86063, 122, 1403, 119, 118,
		118, 1414, 119, 118, 102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(inputIds, expectedInputIds) {
		t.Errorf("Expected '%v', but got '%v'", inputIds, expectedInputIds)
	}

	expectedTokenTypeIds := []int32{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(tokenTypeIds, expectedTokenTypeIds) {
		t.Errorf("Expected '%v', but got '%v'", tokenTypeIds, expectedTokenTypeIds)
	}

	expectedAttentionMask := []int32{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	if !reflect.DeepEqual(attentionMask, expectedAttentionMask) {
		t.Errorf("Expected '%v', but got '%v'", attentionMask, expectedAttentionMask)
	}

	expectedOffsets := []int32{
		0, 1, 1, 3, 3, 4, 4, 5, 5, 6, 7, 8, 8, 12, 12, 13, 13, 15, 15, 16, 16, 17, 17, 18, 19, 20, 20, 21, 21, 22,
		22, 23, 23, 24, 24, 25, 26, 27, 27, 28, 28, 29, 29, 32, 33, 35, 36, 37, 37, 38, 38, 40, 41, 42, 43, 44, 44,
		45, 46, 47, 48, 49, 50, 51, 51, 52, 53, 54,
	}
	if !reflect.DeepEqual(offsets, expectedOffsets) {
		t.Errorf("Expected '%v', but got '%v'", offsets, expectedOffsets)
	}

	tokens, wordPieceOffsets := tk.WordPieceTokenize(testThaiLanguage)

	expectedTokens := []string{
		"น", "##คร", "##ป", "##ฐ", "##ม", "เ", "##มือง", "##น", "##คร", "##ป", "##ฐ", "##ม",
		"ถ", "##น", "##น", "##ข", "##า", "##ด", "เ", "##ล", "##ข", "##ที่", "69", "ห", "##ม",
		"##ู่", "1", "ซ", ".", "-", "-", "ถ", ".", "-",
	}
	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Errorf("Expected '%v', but got '%v'", tokens, expectedTokens)
	}

	expectedWordPieceOffsets := []int32{
		0, 1, 1, 3, 3, 4, 4, 5, 5, 6, 7, 8, 8, 12, 12, 13, 13, 15, 15, 16, 16, 17, 17, 18, 19, 20, 20, 21, 21, 22,
		22, 23, 23, 24, 24, 25, 26, 27, 27, 28, 28, 29, 29, 32, 33, 35, 36, 37, 37, 38, 38, 40, 41, 42, 43, 44, 44,
		45, 46, 47, 48, 49, 50, 51, 51, 52, 53, 54,
	}
	if !reflect.DeepEqual(wordPieceOffsets, expectedWordPieceOffsets) {
		t.Errorf("Expected '%v', but got '%v'", wordPieceOffsets, expectedWordPieceOffsets)
	}

	tk.Close()
}
