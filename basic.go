package tokenizer

import (
	"unicode"
	"unicode/utf8"
)

// Token classes, matching the branch order of the C++ tokenizer.
const (
	classWord  = iota // letter, mark, number or symbol: part of a word
	classCJK          // CJK ideograph: always a token of its own
	classPunct        // punctuation: always a token of its own
	classSpace        // Zs separator: dropped, and ends the current word
	classDrop         // control and unassigned: dropped silently
)

// token is a piece of the input text, covering the byte range [start, end).
//
// text is empty while the token reads exactly like the slice of the input it
// covers, which is the case for most tokens; when the tokenizer rewrites the
// text (lowercasing or normalisation) the rewritten text is stored here.
type token struct {
	start int32
	end   int32
	text  string
}

// tokenText returns the text of tok.
func tokenText(src string, tok token) string {
	if tok.text != "" {
		return tok.text
	}
	if tok.start >= tok.end {
		return ""
	}
	return src[tok.start:tok.end]
}

// appendPiece extends tok with the input range [tok.end, end), whose text is
// piece, and keeps the token backed by the input whenever the pieces still
// read like the input does.
func appendPiece(src string, tok *token, end int32, piece string) {
	if tok.text == "" && piece == src[tok.end:end] {
		tok.end = end
		return
	}
	tok.text = tokenText(src, *tok) + piece
	tok.end = end
}

// mergeWord appends the word to the last token, which happens when a word
// character directly follows a token the tokenizer already emitted.
func mergeWord(src string, tokens []token, word token) []token {
	if len(tokens) == 0 {
		return append(tokens, word)
	}
	last := &tokens[len(tokens)-1]
	appendPiece(src, last, word.end, tokenText(src, word))
	return tokens
}

// flushWord ends the pending word, either by appending it as a new token or by
// extending the token emitted before it. started reports whether a word is
// pending and merge whether it belongs to the previous token.
func flushWord(src string, tokens []token, word token, started, merge bool) []token {
	if !started {
		return tokens
	}
	if !merge {
		return append(tokens, word)
	}
	return mergeWord(src, tokens, word)
}

// maxPrefixMatches bounds how many special tokens are matched at a single
// position, like _max_prefix_matches did in C++.
const maxPrefixMatches = 64

// basicTokenizer splits raw text into words, keeping the special tokens whole.
type basicTokenizer struct {
	doLowerCase bool
	special     *byteTrie
	scan        specialScan
}

func newBasicTokenizer(doLowerCase bool, special SpecialTokens) *basicTokenizer {
	keys := make([]string, 0, 5)
	for _, key := range [5]string{special.Pad, special.CLS, special.SEP, special.UNK, special.Mask} {
		if key != "" {
			keys = append(keys, key)
		}
	}
	return &basicTokenizer{
		doLowerCase: doLowerCase,
		special:     buildTrie(keys),
		scan:        newSpecialScan(keys),
	}
}

// basicTokenize splits text into words. Special tokens are never split, the
// text around them is tokenized normally.
func (b *basicTokenizer) basicTokenize(text string, tokens []token) []token {
	matches := b.specialTokens(text)
	if len(matches) == 0 {
		return b.tokenize(text, 0, len(text), tokens)
	}

	start := 0
	for _, m := range matches {
		if int(m.pos) > start {
			tokens = b.tokenize(text, start, int(m.pos), tokens)
		}
		end := int(m.pos) + len(m.text)
		tokens = append(tokens, token{start: m.pos, end: int32(end), text: m.text})
		start = end
	}
	if start < len(text) {
		tokens = b.tokenize(text, start, len(text), tokens)
	}
	return tokens
}

// specialTokens returns every occurrence of a special token in text, in input
// order.
func (b *basicTokenizer) specialTokens(text string) []specialMatch {
	var out []specialMatch
	for cur := 0; cur < len(text); {
		next := b.scan.next(text[cur:])
		if next < 0 {
			break
		}
		cur += next
		pos := int32(cur)
		b.special.prefixes(text[cur:], maxPrefixMatches, func(id int32) bool {
			out = append(out, specialMatch{pos: pos, text: b.special.keys[id]})
			return true
		})
		cur += charLen(text[cur:])
	}
	return out
}

// specialMatch is one occurrence of a special token in the input.
type specialMatch struct {
	pos  int32
	text string
}

// tokenize splits text[lo:hi] into tokens, appending them to tokens. The
// input is scanned one UTF-8 char at a time, exactly like the C++
// implementation, so the resulting token boundaries are identical.
func (b *basicTokenizer) tokenize(text string, lo, hi int, tokens []token) []token {
	// word is the word currently being scanned, started reports whether it is
	// pending and merge whether it extends the token emitted before it. The
	// C++ implementation decided that when the word began, so the decision is
	// captured there rather than when the word ends.
	var word token
	started := false
	merge := false
	lastState := false
	i := lo

	for i < hi {
		c := text[i]
		if c < utf8.RuneSelf {
			if isASCIIAlnum(c) {
				if !started {
					word = token{start: int32(i), end: int32(i)}
					merge = lastState
					started = true
				}
				n, hasUpper := alnumRun(text[i:hi])
				piece := text[i : i+n]
				if b.doLowerCase && hasUpper {
					piece = lowerASCIIString(piece)
				}
				appendPiece(text, &word, int32(i+n), piece)
				lastState = true
				i += n
				continue
			}

			if !started {
				// A single punctuation, whitespace or control byte.
				c := text[i]
				i++
				if !isCntrlByte(c) {
					lastState = false
					if !isSpaceByte(c) {
						tokens = append(tokens, token{start: int32(i - 1), end: int32(i)})
					}
				}
				continue
			}

			// Flush the pending word and rescan this byte, which then starts
			// the next token.
			tokens = flushWord(text, tokens, word, started, merge)
			started = false
			lastState = true
			continue
		}

		// A non-ASCII char: a leading byte plus its continuation bytes.
		if started {
			tokens = flushWord(text, tokens, word, started, merge)
			started = false
			lastState = true
		}

		n := charLen(text[i:hi])
		chunk := text[i : i+n]
		start := int32(i)
		i += n
		end := int32(i)

		r := decodeChar(chunk)
		switch classifyRune(r) {
		case classCJK, classPunct:
			tokens = append(tokens, token{start: start, end: end})
			lastState = false
		case classSpace:
			lastState = false
		case classDrop:
			// Dropped without changing the state.
		default:
			piece := chunk
			if b.doLowerCase {
				if lower := toLowerRune(r); lower != r {
					piece = encodeRune(lower)
				}
				piece = normalize(piece)
				if len(piece) == 1 && !isASCIIAlnum(piece[0]) && !isCntrlByte(piece[0]) {
					lastState = false
					if !isSpaceByte(piece[0]) {
						tokens = append(tokens, token{start: start, end: end, text: piece})
					}
					continue
				}
			}
			word = token{start: start, end: start}
			merge = lastState
			appendPiece(text, &word, end, piece)
			started = true
			lastState = true
		}
	}

	if started {
		tokens = flushWord(text, tokens, word, started, merge)
	}
	return tokens
}

// classifyRune maps a codepoint to one of the token classes. Invalid UTF-8
// arrives as -1 and is treated like an unassigned codepoint, which is the
// category utf8proc reported for it.
func classifyRune(r rune) int {
	if r < 0 || r == utf8.RuneError {
		return classDrop
	}
	if unicode.IsLetter(r) {
		if isCJK(r) {
			return classCJK
		}
		return classWord
	}
	if unicode.IsNumber(r) || unicode.IsMark(r) || unicode.IsSymbol(r) {
		return classWord
	}
	if unicode.Is(unicode.P, r) {
		return classPunct
	}
	if unicode.Is(unicode.Zs, r) {
		return classSpace
	}
	if unicode.Is(unicode.C, r) || isUnassigned(r) {
		return classDrop
	}
	return classWord
}

// isCJK reports whether r is a CJK ideograph the tokenizer keeps whole.
func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF,
		r >= 0x3400 && r <= 0x4DBF,
		r >= 0x20000 && r <= 0x2A6DF,
		r >= 0x2A700 && r <= 0x2B73F,
		r >= 0x2B740 && r <= 0x2B81F,
		r >= 0x2B820 && r <= 0x2CEAF,
		r >= 0xF900 && r <= 0xFAFF,
		r >= 0x2F800 && r <= 0x2FA1F:
		return true
	}
	return false
}

// isUnassigned reports whether r has no category of its own, i.e. utf8proc's
// category Cn.
func isUnassigned(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsMark(r) && !unicode.IsNumber(r) &&
		!unicode.IsPunct(r) && !unicode.IsSymbol(r) && !unicode.IsSpace(r)
}

// decodeChar decodes one UTF-8 char; invalid input yields -1.
func decodeChar(chunk string) rune {
	r, size := utf8.DecodeRuneInString(chunk)
	if r == utf8.RuneError && size <= 1 {
		return -1
	}
	return r
}

// encodeRune encodes a codepoint as UTF-8.
func encodeRune(r rune) string {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return string(buf[:n])
}
