// Package tokenizer provides a BERT WordPiece tokenizer implemented in pure Go.
//
// The exported API ([NewTokenizer], [EasyTokenizer.Close],
// [EasyTokenizer.Encode], [EasyTokenizer.EncodeWithIds] and
// [EasyTokenizer.WordPieceTokenize]) has been stable since v0.1.0 and is
// unchanged by the v0.3.0 rewrite, which replaced the original C++ backing
// library with this Go implementation.
//
// When the module is built with GOEXPERIMENT=simd the byte and codepoint
// scanning primitives used by the hot paths are vectorised with the standard
// library [simd] package; otherwise a scalar implementation with identical
// semantics is used. Both builds produce the same tokens.
package tokenizer

// Version is the semantic version of this package.
const Version = "0.3.0"

// EasyTokenizer holds a vocabulary and tokenizes text with the WordPiece
// algorithm used by BERT models.
//
// A tokenizer is read-only once created, so a single instance may be used
// concurrently from multiple goroutines.
type EasyTokenizer struct {
	vocab          *vocab
	basic          *basicTokenizer
	codePointLevel bool

	// special is the special token configuration, fixed at construction like
	// everything else the tokenizer holds.
	special SpecialTokens

	clsID  int32
	sepID  int32
	unkID  int32
	padID  int32
	maskID int32

	// specialIDs holds the id of every special token, resolved with the UNK id
	// as the fallback: it is what a special token that is not part of the
	// vocabulary maps to.
	specialIDs []int32
}

// NewTokenizer creates a tokenizer from a vocabulary file, which must contain
// one token per line.
//
// It panics if the vocabulary file cannot be read.
func NewTokenizer(vocabPath string, doLowerCase bool) *EasyTokenizer {
	t, err := NewTokenizerWithOptions(vocabPath, doLowerCase)
	if err != nil {
		panic("tokenizer: cannot load vocabulary " + vocabPath + ": " + err.Error())
	}
	return t
}

// Close releases the tokenizer, and does nothing. It is kept so that code
// written against the v0.1.0 API keeps compiling: the Go implementation holds
// no handle that must be released, and a tokenizer that has been closed stays
// usable, so Close is safe to call at any time, more than once, and
// concurrently with Encode.
func (t *EasyTokenizer) Close() {}

// OffsetType returns the unit of the offsets reported by
// [EasyTokenizer.EncodeWithIds] and [EasyTokenizer.WordPieceTokenize], which
// is always [OffsetCodepoint]. Use [ByteOffset] to turn such an offset into a
// byte offset for slicing the input.
func (t *EasyTokenizer) OffsetType() OffsetType {
	if t.codePointLevel {
		return OffsetCodepoint
	}
	return OffsetByte
}

// Encode tokenizes text and returns a fixed size sequence of token ids.
//
// The result always has maxSeqLength elements: the sequence is truncated to
// maxSeqLength ids and the remaining elements are zero. Special tokens are
// added by default and sequences longer than maxSeqLength are truncated.
//
// Encode is the v0.1.0 entry point. [EasyTokenizer.EncodeWithOptions] returns
// the whole structured result, accepts sentence pairs and batches, and can be
// configured not to pad.
func (t *EasyTokenizer) Encode(text string, maxSeqLength int) []int32 {
	output := make([]int32, maxSeqLength)
	t.fillIds(output, t.wordpieceWalk(text, wpTokenIds).ids)
	return output
}

// EncodeWithIds tokenizes text and returns input ids, token type ids, the
// attention mask and codepoint level offsets.
//
// All of the returned slices have maxSeqLength elements, except for offsets
// which has two entries per token. Special tokens are added by default,
// sequences longer than maxSeqLength are truncated and the result is padded
// with [PAD] ids up to maxSeqLength.
//
// The offsets are codepoint offsets, the unit the original C++ implementation
// used and the value of [EasyTokenizer.OffsetType]; [ByteOffset] converts one
// to a byte offset of text.
func (t *EasyTokenizer) EncodeWithIds(text string, maxSeqLength int) ([]int32, []int32, []int32, []int32) {
	walk := t.wordpieceWalk(text, wpTokenIds)

	inputIds := make([]int32, maxSeqLength)
	length, truncated := t.fillIds(inputIds, walk.ids)
	offsets := walk.offsets
	if truncated {
		// Each token carries a pair of offsets, and the two special tokens
		// do not.
		n := 2*maxSeqLength - 4
		if n < 0 || n > len(offsets) {
			n = len(offsets)
		}
		offsets = offsets[:n]
	}

	// Pad up to maxSeqLength, keeping the attention mask zero beyond the
	// real sequence.
	tokenTypeIds := make([]int32, maxSeqLength)
	attentionMask := make([]int32, maxSeqLength)
	for i := 0; i < length && i < maxSeqLength; i++ {
		attentionMask[i] = 1
	}
	return inputIds, tokenTypeIds, attentionMask, offsets
}

// WordPieceTokenize splits text into WordPiece tokens, prefixed by "##" when
// a token is a continuation of the previous one, and returns the codepoint
// level offsets of every token as (start, end) pairs.
//
// The offsets are codepoint offsets; [ByteOffset] converts one to a byte
// offset of text.
func (t *EasyTokenizer) WordPieceTokenize(text string) ([]string, []int32) {
	walk := t.wordpieceWalk(text, wpTokenTexts)
	return walk.texts, walk.offsets
}

// fillIds writes ids into out, wrapping them in the CLS and SEP ids and padding
// is left to the caller: out is exactly the slot sequence the caller wants to
// return. The walk resolved every id, unknown tokens included, so this is a
// plain copy. When the sequence does not fit, it is truncated and the last slot
// is replaced by SEP, which is what the C++ implementation did. length is the
// number of ids that were written, including that replacement.
func (t *EasyTokenizer) fillIds(out []int32, ids []int32) (length int, truncated bool) {
	if len(out) == 0 {
		// Not even the CLS and SEP ids fit.
		return 0, true
	}
	out[0] = t.clsID
	n := 1
	for _, id := range ids {
		if n == len(out) {
			break
		}
		out[n] = id
		n++
	}
	if n < len(out) {
		out[n] = t.sepID
		return n + 1, false
	}
	out[len(out)-1] = t.sepID
	return len(out), true
}
