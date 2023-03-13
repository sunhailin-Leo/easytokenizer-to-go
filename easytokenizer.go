package tokenizer

// #cgo CXXFLAGS: -std=c++11
// #cgo LDFLAGS: -L${SRCDIR} -ltokenizer -lm
// #include "easytokenizer_wrapper.h"
import "C"
import (
	"unsafe"
)

type EasyTokenizer struct {
	easyTokenizer  C.EasyTokenizer
	vocabPath      string
	doLowerCase    bool
	codePointLevel bool
}

// NewTokenizer init EasyTokenizer
func NewTokenizer(vocabPath string, doLowerCase bool) *EasyTokenizer {
	var tokenizer EasyTokenizer
	tokenizer.vocabPath = vocabPath
	tokenizer.doLowerCase = doLowerCase
	// just true
	tokenizer.codePointLevel = true
	// initTokenizer
	tokenizer.easyTokenizer = C.initTokenizer(C.CString(vocabPath), C.bool(tokenizer.doLowerCase), C.bool(tokenizer.codePointLevel))
	return &tokenizer
}

func (t *EasyTokenizer) Close() {
	C.free(unsafe.Pointer(t.easyTokenizer))
	t.easyTokenizer = nil
}

func (t *EasyTokenizer) Encode(text string, maxSeqLength int) []int32 {
	// text
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	// make allocation of []int32 slice
	outputData := make([]int32, maxSeqLength)
	// call function
	C.encode(t.easyTokenizer, cText, C.bool(true), C.bool(true), C.int(maxSeqLength), (*C.int)(unsafe.Pointer(&outputData[0])))
	return outputData
}

func (t *EasyTokenizer) EncodeWithIds(text string, maxSeqLength int) ([]int32, []int32, []int32, []int32) {
	// text
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	// inputIds, tokenTypeIds, attentionMask, offsets
	var inputIds, tokenTypeIds, attentionMask, offsets *C.int
	// slice number of inputIds, tokenTypeIds, attentionMask, offsets
	var numInputIds, numTokenTypeIds, numAttentionMask, numOffsets C.int
	// call function
	C.encodeWithIds(t.easyTokenizer, cText,
		&inputIds, &numInputIds,
		&tokenTypeIds, &numTokenTypeIds,
		&attentionMask, &numAttentionMask,
		&offsets, &numOffsets,
		C.bool(true), C.bool(true), C.bool(true), C.int(maxSeqLength))
	// to Golang Slice
	sliceInputIds := (*[1 << 30]int32)(unsafe.Pointer(inputIds))[:numInputIds:numInputIds]
	sliceTokenTypeIds := (*[1 << 30]int32)(unsafe.Pointer(tokenTypeIds))[:numTokenTypeIds:numTokenTypeIds]
	sliceAttentionMask := (*[1 << 30]int32)(unsafe.Pointer(attentionMask))[:numAttentionMask:numAttentionMask]
	sliceOffsets := (*[1 << 30]int32)(unsafe.Pointer(offsets))[:numOffsets:numOffsets]
	// release
	defer C.free(unsafe.Pointer(inputIds))
	defer C.free(unsafe.Pointer(tokenTypeIds))
	defer C.free(unsafe.Pointer(attentionMask))
	defer C.free(unsafe.Pointer(offsets))
	return sliceInputIds, sliceTokenTypeIds, sliceAttentionMask, sliceOffsets
}

func (t *EasyTokenizer) WordPieceTokenize(text string) ([]string, []int32) {
	// text
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	// init variable
	var tokens **C.char
	var offsets *C.int
	var numTokens, numOffsets C.int
	// call function
	C.wordPieceTokenize(t.easyTokenizer, cText, &tokens, &numTokens, &offsets, &numOffsets)
	defer C.free(unsafe.Pointer(tokens))
	defer C.free(unsafe.Pointer(offsets))
	// to Golang Slice
	sliceTokens := (*[1 << 30]*C.char)(unsafe.Pointer(tokens))[:numTokens:numTokens]
	sliceOffsets := (*[1 << 30]int32)(unsafe.Pointer(offsets))[:numOffsets:numOffsets]
	// parse string token
	goTokens := make([]string, numTokens)
	for i := 0; i < int(numTokens); i++ {
		goTokens[i] = C.GoString(sliceTokens[i])
	}
	sliceTokens = nil
	return goTokens, sliceOffsets
}
