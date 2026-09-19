package tokenizer

// ByteOffset returns the byte offset in text of a codepoint offset, so that
// text[:ByteOffset(text, start)] is the part of text the offset refers to. The
// tokenizer reports codepoint offsets ([OffsetCodepoint]); this turns one into
// the byte offset to slice with.
//
// Offsets below zero clamp to 0 and offsets past the end of text clamp to
// len(text). An offset that lands inside a multi-byte codepoint rounds down to
// its first byte.
func ByteOffset(text string, codepointOffset int32) int {
	if codepointOffset <= 0 {
		return 0
	}
	i := 0
	for n := int32(0); n < codepointOffset && i < len(text); n++ {
		i += charLen(text[i:])
	}
	return i
}

// TruncationStrategy selects which segment of a sentence pair is shortened
// when the encoding does not fit the length limit.
type TruncationStrategy int

const (
	// TruncateLongestFirst drops tokens from the longer segment first, which is
	// what BERT pre-processing does in practice. Ties shorten the second
	// segment, so a query stays whole when the two halves are the same length.
	TruncateLongestFirst TruncationStrategy = iota
	// TruncateOnlyFirst drops tokens from the first segment, and falls back to
	// the tail of the second when the first one is exhausted.
	TruncateOnlyFirst
	// TruncateOnlySecond drops tokens from the second segment, and falls back
	// to the tail of the first when the second one is exhausted.
	TruncateOnlySecond
)

// PaddingStrategy selects whether an encoding is padded to the length limit.
type PaddingStrategy int

const (
	// PadToMaxLength pads every encoding to MaxSeqLength with the PAD id. It is
	// the default.
	PadToMaxLength PaddingStrategy = iota
	// PadNone returns exactly as many slots as there are tokens.
	PadNone
)

// EncodingOptions configures [EasyTokenizer.EncodeWithOptions],
// [EasyTokenizer.EncodePairWithOptions] and
// [EasyTokenizer.EncodeBatchWithOptions].
//
// MaxSeqLength is the truncation limit and the padding target at the same
// time. Zero or less means no limit and no padding, so the result holds
// exactly the tokens of the input plus the special tokens. A limit of 1
// returns just [CLS] and a limit of 2 returns [CLS] [SEP]; the strategies
// below apply from 3 upwards.
//
// Truncation applies to sentence pairs: a single text is always shortened
// from the tail. Padding selects whether the result is padded to MaxSeqLength
// with the PAD id of the tokenizer.
//
// The zero value is valid and encodes the full text without padding.
type EncodingOptions struct {
	MaxSeqLength int
	Truncation   TruncationStrategy
	Padding      PaddingStrategy
}

// Encoding is the structured result of an encoding call.
//
// IDs, TokenTypeIDs, AttentionMask and Tokens have the same length: the
// sequence length, which is MaxSeqLength when the encoding is padded and the
// exact token count otherwise. Tokens holds the WordPiece text of every slot,
// including the special tokens, with continuation pieces prefixed by "##".
//
// Offsets holds the two offsets of every token that came from the input text.
// Special tokens and padding have no offsets, so len(Offsets) is twice the
// number of tokens that did come from the input. The offsets of a sentence
// pair are relative to the sentence each token came from.
type Encoding struct {
	IDs           []int32
	TokenTypeIDs  []int32
	AttentionMask []int32
	Tokens        []string
	Offsets       []int32
}

// EncodeWithOptions tokenizes text and returns the whole structured result.
//
// Unlike Encode, the result is not fixed size unless EncodingOptions asks for
// it: with a MaxSeqLength of zero it holds the full sequence, and with
// [PadNone] it is not padded to the limit. Padding uses the PAD id of the
// tokenizer, which is what the frozen Encode and EncodeWithIds cannot do
// because they fill with zero.
func (t *EasyTokenizer) EncodeWithOptions(text string, opts EncodingOptions) Encoding {
	walk := t.wordpieceWalk(text, wpTokenAll)
	part := seqPart{ids: walk.ids, texts: walk.texts, offsets: walk.offsets}
	return t.assemble([]seqPart{part.truncated(t.contentBudget(opts.MaxSeqLength))}, opts)
}

// EncodePair tokenizes the sentence pair (a, b) into a fixed size sequence of
// token ids.
//
// The sequence is [CLS] a [SEP] b [SEP], padded with the PAD id up to
// maxSeqLength. A maxSeqLength of zero returns the full sequence without
// padding. Truncation drops tokens from the longer of the two sentences first;
// [EasyTokenizer.EncodePairWithOptions] selects another strategy.
func (t *EasyTokenizer) EncodePair(a, b string, maxSeqLength int) []int32 {
	return t.EncodePairWithOptions(a, b, EncodingOptions{MaxSeqLength: maxSeqLength}).IDs
}

// EncodePairWithIds tokenizes the sentence pair (a, b) and returns input ids,
// token type ids, the attention mask and codepoint level offsets, mirroring
// [EasyTokenizer.EncodeWithIds].
//
// Token type ids are 0 for [CLS], the first sentence and the [SEP] after it,
// and 1 for the second sentence and the [SEP] that closes it. The offsets
// cover the tokens that came from a and b, not the special tokens, and are
// relative to the sentence each token came from.
func (t *EasyTokenizer) EncodePairWithIds(a, b string, maxSeqLength int) ([]int32, []int32, []int32, []int32) {
	enc := t.EncodePairWithOptions(a, b, EncodingOptions{MaxSeqLength: maxSeqLength})
	return enc.IDs, enc.TokenTypeIDs, enc.AttentionMask, enc.Offsets
}

// EncodePairWithOptions tokenizes the sentence pair (a, b) and returns the
// whole structured result, with the truncation and padding strategy of opts.
func (t *EasyTokenizer) EncodePairWithOptions(a, b string, opts EncodingOptions) Encoding {
	wa := t.wordpieceWalk(a, wpTokenAll)
	wb := t.wordpieceWalk(b, wpTokenAll)
	parts := []seqPart{
		{ids: wa.ids, texts: wa.texts, offsets: wa.offsets},
		{ids: wb.ids, texts: wb.texts, offsets: wb.offsets, typeID: 1},
	}
	if limit := opts.MaxSeqLength; limit >= 3 {
		parts = truncatePair(parts, limit-3, opts.Truncation)
	}
	return t.assemble(parts, opts)
}

// EncodeBatch tokenizes every text of the batch and returns their id sequences,
// one per input text, exactly as a loop over [EasyTokenizer.Encode] would.
func (t *EasyTokenizer) EncodeBatch(texts []string, maxSeqLength int) [][]int32 {
	out := make([][]int32, len(texts))
	for i, text := range texts {
		out[i] = t.Encode(text, maxSeqLength)
	}
	return out
}

// EncodeBatchWithIds tokenizes every text of the batch and returns the four
// slices of [EasyTokenizer.EncodeWithIds] for each of them.
func (t *EasyTokenizer) EncodeBatchWithIds(texts []string, maxSeqLength int) (ids, tokenTypeIds, attentionMasks, offsets [][]int32) {
	ids = make([][]int32, len(texts))
	tokenTypeIds = make([][]int32, len(texts))
	attentionMasks = make([][]int32, len(texts))
	offsets = make([][]int32, len(texts))
	for i, text := range texts {
		ids[i], tokenTypeIds[i], attentionMasks[i], offsets[i] = t.EncodeWithIds(text, maxSeqLength)
	}
	return ids, tokenTypeIds, attentionMasks, offsets
}

// EncodeBatchWithOptions tokenizes every text of the batch with the same
// options and returns one structured result per input text.
func (t *EasyTokenizer) EncodeBatchWithOptions(texts []string, opts EncodingOptions) []Encoding {
	out := make([]Encoding, len(texts))
	for i, text := range texts {
		out[i] = t.EncodeWithOptions(text, opts)
	}
	return out
}

// contentBudget returns how many content tokens fit next to the [CLS] and
// [SEP] of a single sequence, or -1 when there is no limit.
func (t *EasyTokenizer) contentBudget(maxSeqLength int) int {
	if maxSeqLength <= 0 {
		return -1
	}
	if maxSeqLength < 3 {
		return 0
	}
	return maxSeqLength - 2
}

// truncatePair drops tokens from the two segments until they fit budget,
// following the strategy. budget is never negative.
func truncatePair(parts []seqPart, budget int, strategy TruncationStrategy) []seqPart {
	for len(parts[0].ids)+len(parts[1].ids) > budget {
		switch strategy {
		case TruncateOnlyFirst:
			if len(parts[0].ids) > 0 {
				parts[0].dropLast()
				continue
			}
			parts[1].dropLast()
		case TruncateOnlySecond:
			if len(parts[1].ids) > 0 {
				parts[1].dropLast()
				continue
			}
			parts[0].dropLast()
		default:
			if len(parts[0].ids) > len(parts[1].ids) {
				parts[0].dropLast()
			} else {
				parts[1].dropLast()
			}
		}
	}
	return parts
}

// seqPart is one segment of an encoding: the tokens of a text, their ids and
// their flat offset pairs. typeID is the value written to the token type id
// sequence of the segment and of the [SEP] that closes it.
type seqPart struct {
	ids     []int32
	texts   []string
	offsets []int32
	typeID  int32
}

// truncated returns the part with at most n tokens, or the whole part when n
// is negative.
func (p seqPart) truncated(n int) seqPart {
	if n < 0 || len(p.ids) <= n {
		return p
	}
	if n == 0 {
		return seqPart{typeID: p.typeID}
	}
	p.ids = p.ids[:n]
	p.texts = p.texts[:n]
	p.offsets = p.offsets[:2*n]
	return p
}

// dropLast removes the last token of the part, which must not be empty.
func (p *seqPart) dropLast() {
	n := len(p.ids) - 1
	p.ids = p.ids[:n]
	p.texts = p.texts[:n]
	p.offsets = p.offsets[:2*n]
}

// assemble writes the parts as [CLS] part0 [SEP] part1 [SEP], pads the result
// to the length limit of opts with the PAD id and truncates it. The callers
// have already dropped the tokens that do not fit, so the assembled sequence
// never exceeds the limit.
func (t *EasyTokenizer) assemble(parts []seqPart, opts EncodingOptions) Encoding {
	limit := opts.MaxSeqLength
	if limit > 0 && limit < 3 {
		return t.onlySpecialEncoding(limit)
	}

	real := 1 + len(parts)
	for _, p := range parts {
		real += len(p.ids)
	}
	length := real
	if limit > 0 {
		if length > limit {
			length = limit
		}
		if length < limit && opts.Padding != PadNone {
			length = limit
		}
	}

	padID := t.padID
	if padID < 0 {
		padID = 0
	}
	enc := Encoding{
		IDs:           make([]int32, length),
		TokenTypeIDs:  make([]int32, length),
		AttentionMask: make([]int32, length),
		Tokens:        make([]string, length),
		Offsets:       make([]int32, 0, 2*real),
	}
	for i := range enc.IDs {
		enc.IDs[i] = padID
		enc.Tokens[i] = t.special.Pad
	}

	at := 0
	t.writeSlot(&enc, &at, t.clsID, t.special.CLS, 0)
	for _, p := range parts {
		for k, id := range p.ids {
			t.writeSlot(&enc, &at, id, p.texts[k], p.typeID)
			enc.Offsets = append(enc.Offsets, p.offsets[2*k], p.offsets[2*k+1])
		}
		t.writeSlot(&enc, &at, t.sepID, t.special.SEP, p.typeID)
	}
	return enc
}

// onlySpecialEncoding returns the degenerate [CLS] / [CLS] [SEP] result of a
// length limit below three.
func (t *EasyTokenizer) onlySpecialEncoding(length int) Encoding {
	enc := Encoding{
		IDs:           make([]int32, length),
		TokenTypeIDs:  make([]int32, length),
		AttentionMask: make([]int32, length),
		Tokens:        make([]string, length),
	}
	at := 0
	t.writeSlot(&enc, &at, t.clsID, t.special.CLS, 0)
	if length == 2 {
		t.writeSlot(&enc, &at, t.sepID, t.special.SEP, 0)
	}
	return enc
}

// writeSlot writes one real slot at *at and advances it.
func (t *EasyTokenizer) writeSlot(enc *Encoding, at *int, id int32, text string, typeID int32) {
	enc.IDs[*at] = id
	enc.Tokens[*at] = text
	enc.TokenTypeIDs[*at] = typeID
	enc.AttentionMask[*at] = 1
	*at++
}
