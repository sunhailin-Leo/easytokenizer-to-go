package main

import (
	"errors"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	tokenizer "github.com/sunhailin-Leo/easytokenizer-to-go"
)

// smallVocab is a hand written vocabulary with predictable ids:
// [PAD]=0 [UNK]=1 [CLS]=2 [SEP]=3 [MASK]=4 one=5 ... six=10.
func smallVocab() fstest.MapFS {
	return fstest.MapFS{
		"vocab.txt": &fstest.MapFile{Data: []byte(strings.Join([]string{
			"[PAD]", "[UNK]", "[CLS]", "[SEP]", "[MASK]",
			"one", "two", "three", "four", "five", "six",
		}, "\n") + "\n")},
	}
}

func TestEncodingOptionsMatchFrozenAPI(t *testing.T) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testMultilingualVocabFileName, false)
	defer tk.Close()

	const maxSeq = 64
	enc := tk.EncodeWithOptions(testThaiLanguage, tokenizer.EncodingOptions{MaxSeqLength: maxSeq})
	inputIds, typeIds, mask, offsets := tk.EncodeWithIds(testThaiLanguage, maxSeq)

	if !reflect.DeepEqual(enc.IDs, tk.Encode(testThaiLanguage, maxSeq)) {
		t.Errorf("IDs differ from Encode: %v", enc.IDs)
	}
	if !reflect.DeepEqual(enc.IDs, inputIds) {
		t.Errorf("IDs differ from EncodeWithIds: %v", enc.IDs)
	}
	if !reflect.DeepEqual(enc.TokenTypeIDs, typeIds) {
		t.Errorf("TokenTypeIDs differ from EncodeWithIds: %v", enc.TokenTypeIDs)
	}
	if !reflect.DeepEqual(enc.AttentionMask, mask) {
		t.Errorf("AttentionMask differ from EncodeWithIds: %v", enc.AttentionMask)
	}
	if !reflect.DeepEqual(enc.Offsets, offsets) {
		t.Errorf("Offsets differ from EncodeWithIds: %v", enc.Offsets)
	}

	tokens, tokOffsets := tk.WordPieceTokenize(testThaiLanguage)
	if !reflect.DeepEqual(enc.Offsets, tokOffsets) {
		t.Errorf("Offsets differ from WordPieceTokenize: %v", enc.Offsets)
	}
	wantTokens := append([]string{"[CLS]"}, tokens...)
	wantTokens = append(wantTokens, "[SEP]")
	if !reflect.DeepEqual(enc.Tokens[:len(wantTokens)], wantTokens) {
		t.Errorf("Tokens = %v, want %v", enc.Tokens, wantTokens)
	}
	if len(enc.Tokens) != maxSeq || enc.Tokens[len(wantTokens)] != "[PAD]" {
		t.Errorf("padding tokens = %v", enc.Tokens[len(wantTokens):])
	}
}

func TestByteOffset(t *testing.T) {
	text := "广东省abcé"
	// The byte offsets of the codepoint starts, computed independently.
	want := []int{0, 3, 6, 9, 10, 11, 12}
	for cp, byteOffset := range want {
		if got := tokenizer.ByteOffset(text, int32(cp)); got != byteOffset {
			t.Errorf("ByteOffset(%q, %d) = %d, want %d", text, cp, got, byteOffset)
		}
	}
	if got := tokenizer.ByteOffset(text, -3); got != 0 {
		t.Errorf("ByteOffset with a negative offset = %d, want 0", got)
	}
	if got := tokenizer.ByteOffset(text, 1000); got != len(text) {
		t.Errorf("ByteOffset past the end = %d, want %d", got, len(text))
	}
}

func TestOffsetType(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromFS(smallVocab(), "vocab.txt", false)
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}
	if got := tk.OffsetType(); got != tokenizer.OffsetCodepoint {
		t.Errorf("OffsetType() = %v, want OffsetCodepoint", got)
	}
}

func TestNewTokenizerFromReader(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromReader(strings.NewReader("[CLS]\nhello\n[SEP]\nworld\n"), false)
	if err != nil {
		t.Fatalf("NewTokenizerFromReader: %v", err)
	}
	ids := tk.Encode("hello", 4)
	if !reflect.DeepEqual(ids, []int32{0, 1, 2, 0}) {
		t.Errorf("Encode(hello, 4) = %v, want [0 1 2 0]", ids)
	}
}

func TestConstructorErrors(t *testing.T) {
	if _, err := tokenizer.NewTokenizerWithOptions("/nonexistent/vocab.txt", true); err == nil {
		t.Error("NewTokenizerWithOptions with a missing file should return an error")
	}
	if _, err := tokenizer.NewTokenizerFromFS(smallVocab(), "missing.txt", true); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("NewTokenizerFromFS with a missing path = %v, want fs.ErrNotExist", err)
	}
}

func TestCloseIsNoOp(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromFS(smallVocab(), "vocab.txt", false)
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}
	before := tk.Encode("one two", 8)
	tk.Close()
	if got := tk.Encode("one two", 8); !reflect.DeepEqual(got, before) {
		t.Errorf("Encode after Close = %v, want %v", got, before)
	}
}

func TestCustomSpecialTokens(t *testing.T) {
	vocab := fstest.MapFS{
		"vocab.txt": &fstest.MapFile{Data: []byte("[PAD]\n<s>\n</s>\n<unk>\n[MASK]\nhi\n")},
	}
	tk, err := tokenizer.NewTokenizerFromFS(vocab, "vocab.txt", false,
		tokenizer.WithSpecialTokens(tokenizer.SpecialTokens{
			Pad: "[PAD]", CLS: "<s>", SEP: "</s>", UNK: "<unk>", Mask: "[MASK]",
		}))
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}

	tokens, _ := tk.WordPieceTokenize("<s> hi </s> [PAD]")
	want := []string{"<s>", "hi", "</s>", "[PAD]"}
	if !reflect.DeepEqual(tokens, want) {
		t.Errorf("WordPieceTokenize = %v, want %v", tokens, want)
	}

	ids := tk.Encode("hi", 5)
	if !reflect.DeepEqual(ids, []int32{1, 5, 2, 0, 0}) {
		t.Errorf("Encode = %v, want [1 5 2 0 0]", ids)
	}

	withoutMask, err := tokenizer.NewTokenizerFromFS(vocab, "vocab.txt", false,
		tokenizer.WithSpecialTokens(tokenizer.SpecialTokens{Pad: "[PAD]", CLS: "<s>", SEP: "</s>", UNK: "<unk>"}))
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}
	if tokens, _ := withoutMask.WordPieceTokenize("[MASK]"); len(tokens) == 1 {
		t.Errorf("a removed [MASK] token should not be kept whole: %v", tokens)
	}
}

func TestEncodePair(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromFS(smallVocab(), "vocab.txt", false)
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}

	// three tokens each, limit 8: the budget is 5, so a three token tie
	// shortens the second sentence.
	ids, typeIds, mask, offsets := tk.EncodePairWithIds("one two three", "four five six", 8)
	wantIDs := []int32{2, 5, 6, 7, 3, 8, 9, 3}
	wantTypes := []int32{0, 0, 0, 0, 0, 1, 1, 1}
	wantMask := []int32{1, 1, 1, 1, 1, 1, 1, 1}
	if !reflect.DeepEqual(ids, wantIDs) {
		t.Errorf("EncodePair ids = %v, want %v", ids, wantIDs)
	}
	if !reflect.DeepEqual(typeIds, wantTypes) {
		t.Errorf("EncodePair type ids = %v, want %v", typeIds, wantTypes)
	}
	if !reflect.DeepEqual(mask, wantMask) {
		t.Errorf("EncodePair mask = %v, want %v", mask, wantMask)
	}
	if len(offsets) != 2*5 {
		t.Errorf("EncodePair offsets = %v, want 5 token pairs", offsets)
	}
	if !reflect.DeepEqual(tk.EncodePair("one two three", "four five six", 8), wantIDs) {
		t.Errorf("EncodePair = %v, want %v", tk.EncodePair("one two three", "four five six", 8), wantIDs)
	}

	// Without a limit the pair is not truncated and not padded.
	enc := tk.EncodePairWithOptions("one two", "three", tokenizer.EncodingOptions{MaxSeqLength: 0})
	wantIDs = []int32{2, 5, 6, 3, 7, 3}
	if !reflect.DeepEqual(enc.IDs, wantIDs) {
		t.Errorf("EncodePairWithOptions ids = %v, want %v", enc.IDs, wantIDs)
	}
	if len(enc.Offsets) != 2*3 {
		t.Errorf("offsets = %v, want 3 token pairs", enc.Offsets)
	}
	wantTokens := []string{"[CLS]", "one", "two", "[SEP]", "three", "[SEP]"}
	if !reflect.DeepEqual(enc.Tokens, wantTokens) {
		t.Errorf("tokens = %v, want %v", enc.Tokens, wantTokens)
	}

	// PadNone keeps the exact length even with a limit.
	enc = tk.EncodePairWithOptions("one two", "three", tokenizer.EncodingOptions{
		MaxSeqLength: 32,
		Padding:      tokenizer.PadNone,
	})
	if len(enc.IDs) != 6 || len(enc.AttentionMask) != 6 {
		t.Errorf("PadNone ids = %v", enc.IDs)
	}

	// Truncation drops from the longer sentence first; a strategy that empties
	// its own segment falls back to the tail of the other one.
	long := strings.Repeat("one ", 8)
	enc = tk.EncodePairWithOptions(long, "two three", tokenizer.EncodingOptions{
		MaxSeqLength: 9,
		Truncation:   tokenizer.TruncateOnlyFirst,
	})
	if !reflect.DeepEqual(enc.IDs, []int32{2, 5, 5, 5, 5, 3, 6, 7, 3}) {
		t.Errorf("TruncateOnlyFirst ids = %v", enc.IDs)
	}
	enc = tk.EncodePairWithOptions("one two", long, tokenizer.EncodingOptions{
		MaxSeqLength: 9,
		Truncation:   tokenizer.TruncateOnlySecond,
	})
	if !reflect.DeepEqual(enc.IDs, []int32{2, 5, 6, 3, 5, 5, 5, 5, 3}) {
		t.Errorf("TruncateOnlySecond ids = %v", enc.IDs)
	}
	enc = tk.EncodePairWithOptions(long, "two three four", tokenizer.EncodingOptions{
		MaxSeqLength: 5,
		Truncation:   tokenizer.TruncateOnlyFirst,
	})
	if !reflect.DeepEqual(enc.IDs, []int32{2, 3, 6, 7, 3}) {
		t.Errorf("TruncateOnlyFirst fallback ids = %v", enc.IDs)
	}
	enc = tk.EncodePairWithOptions(long, "two three four", tokenizer.EncodingOptions{
		MaxSeqLength: 5,
		Truncation:   tokenizer.TruncateOnlySecond,
	})
	if !reflect.DeepEqual(enc.IDs, []int32{2, 5, 5, 3, 3}) {
		t.Errorf("TruncateOnlySecond fallback ids = %v", enc.IDs)
	}

	// The degenerate limits.
	if ids := tk.EncodePair("one two", "three", 2); !reflect.DeepEqual(ids, []int32{2, 3}) {
		t.Errorf("EncodePair limit 2 = %v, want [2 3]", ids)
	}
	if ids := tk.EncodePair("one two", "three", 1); !reflect.DeepEqual(ids, []int32{2}) {
		t.Errorf("EncodePair limit 1 = %v, want [2]", ids)
	}
}

func TestEncodeBatch(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromFS(smallVocab(), "vocab.txt", false)
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}
	texts := []string{"one two", "three", "", "four five six"}

	batch := tk.EncodeBatch(texts, 6)
	if len(batch) != len(texts) {
		t.Fatalf("EncodeBatch returned %d results, want %d", len(batch), len(texts))
	}
	for i, text := range texts {
		if want := tk.Encode(text, 6); !reflect.DeepEqual(batch[i], want) {
			t.Errorf("EncodeBatch[%d] = %v, want %v", i, batch[i], want)
		}
	}

	ids, typeIds, masks, offsets := tk.EncodeBatchWithIds(texts, 6)
	for i, text := range texts {
		wantIDs, wantTypes, wantMasks, wantOffsets := tk.EncodeWithIds(text, 6)
		if !reflect.DeepEqual(ids[i], wantIDs) || !reflect.DeepEqual(typeIds[i], wantTypes) ||
			!reflect.DeepEqual(masks[i], wantMasks) || !reflect.DeepEqual(offsets[i], wantOffsets) {
			t.Errorf("EncodeBatchWithIds[%d] differs from EncodeWithIds", i)
		}
	}

	structured := tk.EncodeBatchWithOptions(texts, tokenizer.EncodingOptions{MaxSeqLength: 6})
	for i := range texts {
		if !reflect.DeepEqual(structured[i].IDs, ids[i]) {
			t.Errorf("EncodeBatchWithOptions[%d] ids = %v, want %v", i, structured[i].IDs, ids[i])
		}
	}
}

func TestEncodingIsSelfConsistent(t *testing.T) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)
	defer tk.Close()

	enc := tk.EncodeWithOptions("广东省深圳市南山区腾讯滨海大厦", tokenizer.EncodingOptions{MaxSeqLength: 48})
	if len(enc.IDs) != len(enc.Tokens) || len(enc.IDs) != len(enc.TokenTypeIDs) || len(enc.IDs) != len(enc.AttentionMask) {
		t.Fatalf("slices have different lengths: %d %d %d %d",
			len(enc.IDs), len(enc.Tokens), len(enc.TokenTypeIDs), len(enc.AttentionMask))
	}
	real := 0
	for _, m := range enc.AttentionMask {
		if m == 1 {
			real++
		}
	}
	if len(enc.Offsets) != 2*(real-2) {
		t.Errorf("offsets = %d entries for %d real slots (2 of them special)", len(enc.Offsets), real)
	}
	if enc.IDs[0] != 101 || enc.Tokens[0] != "[CLS]" || enc.IDs[real-1] != 102 {
		t.Errorf("sequence does not start with [CLS] and end with [SEP]: %v", enc.Tokens[:real])
	}
}

// TestEncodingLimits walks every small length limit, which is where the slot
// arithmetic of a pair is easiest to get wrong.
func TestEncodingLimits(t *testing.T) {
	tk, err := tokenizer.NewTokenizerFromFS(smallVocab(), "vocab.txt", false)
	if err != nil {
		t.Fatalf("NewTokenizerFromFS: %v", err)
	}
	for limit := 0; limit <= 12; limit++ {
		opts := tokenizer.EncodingOptions{MaxSeqLength: limit}
		enc := tk.EncodeWithOptions("one two three", opts)
		if want := limit; limit > 0 && len(enc.IDs) != want {
			t.Errorf("EncodeWithOptions limit %d has %d slots", limit, len(enc.IDs))
		}
		if len(enc.Tokens) != len(enc.IDs) || len(enc.AttentionMask) != len(enc.IDs) {
			t.Errorf("EncodeWithOptions limit %d slices disagree", limit)
		}

		pair := tk.EncodePairWithOptions("one two three", "four five six", opts)
		if limit > 0 && len(pair.IDs) != limit {
			t.Errorf("EncodePairWithOptions limit %d has %d slots", limit, len(pair.IDs))
		}
		real := 0
		for _, m := range pair.AttentionMask {
			if m == 1 {
				real++
			}
		}
		if limit != 1 && pair.IDs[real-1] != 3 {
			t.Errorf("EncodePairWithOptions limit %d does not end with [SEP]: %v", limit, pair.IDs)
		}
	}
}

// TestEncodingOptionsAgreeWithFrozenEncode pins the promise that the new
// entry point reproduces the frozen Encode for every length limit it accepts
// (0 and 1 mean something else by design: no limit, and CLS only).
func TestEncodingOptionsAgreeWithFrozenEncode(t *testing.T) {
	pwd, _ := os.Getwd()
	tk := tokenizer.NewTokenizer(pwd+"/"+testChineseVocabFilename, true)
	defer tk.Close()

	texts := []string{"", "hi", "广东省深圳市南山区腾讯滨海大厦", strings.Repeat("广东", 30)}
	for _, text := range texts {
		for _, maxSeq := range []int{2, 3, 5, 8, 16, 48} {
			want := tk.Encode(text, maxSeq)
			got := tk.EncodeWithOptions(text, tokenizer.EncodingOptions{MaxSeqLength: maxSeq}).IDs
			if !reflect.DeepEqual(got, want) {
				t.Errorf("EncodeWithOptions(%q, %d) ids = %v, want %v", text, maxSeq, got, want)
			}

			_, _, wantMask, wantOffsets := tk.EncodeWithIds(text, maxSeq)
			enc := tk.EncodeWithOptions(text, tokenizer.EncodingOptions{MaxSeqLength: maxSeq})
			if !reflect.DeepEqual(enc.AttentionMask, wantMask) {
				t.Errorf("EncodeWithOptions(%q, %d) mask = %v, want %v", text, maxSeq, enc.AttentionMask, wantMask)
			}
			if !sameInt32s(enc.Offsets, wantOffsets) {
				t.Errorf("EncodeWithOptions(%q, %d) offsets = %v, want %v", text, maxSeq, enc.Offsets, wantOffsets)
			}
		}
	}
}

// sameInt32s compares two int32 slices, treating nil and an empty slice as the
// same, which reflect.DeepEqual does not.
func sameInt32s(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	return len(a) == 0 || reflect.DeepEqual(a, b)
}
