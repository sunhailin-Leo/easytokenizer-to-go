package tokenizer

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// toLowerRune mirrors utf8proc_tolower: the simple lowercase mapping of a
// codepoint, or the codepoint itself when it has none.
func toLowerRune(r rune) rune {
	if r < utf8.RuneSelf {
		if 'A' <= r && r <= 'Z' {
			return r + 'a' - 'A'
		}
		return r
	}
	return unicode.ToLower(r)
}

// normalize returns the NFD form of s with nonspacing marks (category Mn)
// removed, which is how "é" becomes "e".
func normalize(s string) string {
	return stripMn(norm.NFD.String(s))
}

// nfdCodepointNumber counts the codepoints of the NFD form of s, ignoring
// nonspacing marks.
func nfdCodepointNumber(s string) int {
	n := 0
	for _, r := range norm.NFD.String(s) {
		if r < utf8.RuneSelf {
			n++
			continue
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		n++
	}
	return n
}

// stripMn removes every nonspacing mark from the NFD decomposed input.
func stripMn(s string) string {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if isMn(r) {
			break
		}
		i += size
	}
	if i == len(s) {
		return s
	}

	out := make([]byte, 0, len(s))
	out = append(out, s[:i]...)
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if isMn(r) {
			i += size
			continue
		}
		out = append(out, s[i:i+size]...)
		i += size
	}
	return string(out)
}

// isMn reports whether r is a nonspacing mark. ASCII never is.
func isMn(r rune) bool {
	return r >= utf8.RuneSelf && unicode.Is(unicode.Mn, r)
}
