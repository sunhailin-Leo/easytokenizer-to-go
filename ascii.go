package tokenizer

import "strings"

// ASCII byte helpers. The C++ implementation used the C locale functions
// isalnum/isspace/iscntrl, whose behaviour on bytes below 0x80 is fixed.

// isASCIIAlnum reports whether c is in [0-9A-Za-z].
func isASCIIAlnum(c byte) bool {
	return '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
}

// isSpaceByte reports whether c is ASCII whitespace.
func isSpaceByte(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	return false
}

// isCntrlByte reports whether c is an ASCII control character, except for
// '\t', '\r' and '\n' which the tokenizer treats as whitespace instead.
func isCntrlByte(c byte) bool {
	switch c {
	case '\t', '\r', '\n':
		return false
	}
	return c < 0x20 || c == 0x7F
}

// scalarAlnumRun returns the number of leading ASCII alphanumeric bytes of s
// and whether any of them is an uppercase letter.
func scalarAlnumRun(s string) (int, bool) {
	return scalarAlnumRunBudget(s, len(s))
}

// scalarAlnumRunBudget is scalarAlnumRun that gives up after limit bytes.
func scalarAlnumRunBudget(s string, limit int) (int, bool) {
	if limit > len(s) {
		limit = len(s)
	}
	hasUpper := false
	i := 0
	for i < limit {
		c := s[i]
		if !isASCIIAlnum(c) {
			break
		}
		if 'A' <= c && c <= 'Z' {
			hasUpper = true
		}
		i++
	}
	return i, hasUpper
}

// allASCIIAlnum reports whether every byte of s is alphanumeric ASCII.
//
// The predicate is always evaluated with the scalar loop: it is only ever
// called on single words, and a vector kernel without an early exit is far
// slower on the words that fail on a punctuation byte near the front.
func allASCIIAlnum(s string) bool {
	for i := 0; i < len(s); i++ {
		if !isASCIIAlnum(s[i]) {
			return false
		}
	}
	return true
}

// scalarCountCodepoints returns the number of UTF-8 encoded codepoints in s,
// where a codepoint is a byte that is not a continuation byte.
func scalarCountCodepoints(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i]&0xC0 != 0x80 {
			n++
		}
	}
	return n
}

// lowerASCIIString returns s with its ASCII letters lowercased, returning s
// itself when there is nothing to do.
func lowerASCIIString(s string) string {
	for i := 0; i < len(s); i++ {
		if 'A' <= s[i] && s[i] <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if 'A' <= b[j] && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return s
}

// nextSpecialStart returns the offset of the next b byte of s, or -1. It is
// how the tokenizer finds the first byte of a special token, and it is used
// when every special token is spelled with the same leading byte.
//
// The scan is stdlib: strings.IndexByte has a hand written NEON kernel on
// arm64 that reduces its comparison mask to a byte index in hardware. The
// portable simd package cannot express that reduction on arm64, so a
// hand-written vector loop here measured slower than the plain scalar loop,
// which in turn measured ~20x slower than IndexByte.
func nextSpecialStart(s string, b byte) int {
	return strings.IndexByte(s, b)
}
