package tokenizer

import "golang.org/x/text/unicode/norm"

// npos marks the byte positions inside a multi-byte codepoint in the byte to
// codepoint index map. It is the int32 view of the SizeT(-1) the C++
// implementation used.
const npos = int32(-1)

// charLen returns the length of the UTF-8 char starting at s[0]: a leading
// byte followed by its continuation bytes. Invalid sequences are measured
// with the same rule the C++ implementation used.
func charLen(s string) int {
	if len(s) == 0 {
		return 0
	}
	i := 1
	for i < len(s) && s[i]&0xC0 == 0x80 {
		i++
	}
	return i
}

// buildIndexMap maps every byte offset of text to the codepoint index that
// starts there; the byte offsets inside a codepoint map to npos. The final
// entry holds the codepoint count.
func buildIndexMap(text string) []int32 {
	m := make([]int32, len(text)+1)
	for i := range m {
		m[i] = npos
	}
	cur, index := 0, int32(0)
	for cur < len(text) {
		m[cur] = index
		index++
		cur += charLen(text[cur:])
	}
	m[cur] = index
	return m
}

// getCodepointNumber returns the number of UTF-8 encoded codepoints in s.
func getCodepointNumber(s string) int {
	return countCodepoints(s)
}

// search returns the byte offset of codepoint index inside s, or npos when s
// holds fewer codepoints.
func search(s string, index int) int {
	cur, i := 0, 0
	for cur < len(s) {
		if i == index {
			return cur
		}
		cur += charLen(s[cur:])
		i++
	}
	if i == index {
		return len(s)
	}
	return int(npos)
}

// buildPosMap appends, for every codepoint of s, the index of the byte it
// starts at, honouring the lowercase normalisation the tokenizer applies.
func buildPosMap(s string, doLowerCase bool, posMap []int32) []int32 {
	val := int32(0)
	for cur := 0; cur < len(s); {
		c := s[cur]
		if c < 0x80 {
			if isCntrlByte(c) {
				val++
				cur++
				continue
			}
			posMap = append(posMap, val)
			val++
			cur++
			continue
		}

		n := charLen(s[cur:])
		r := decodeChar(s[cur : cur+n])
		cur += n

		if classifyRune(r) == classDrop {
			val++
			continue
		}
		if doLowerCase {
			for i := nfdCodepointNumber(encodeRune(toLowerRune(r))); i > 0; i-- {
				posMap = append(posMap, val)
			}
		} else {
			posMap = append(posMap, val)
		}
		val++
	}
	return append(posMap, val)
}

// posMapIdentity reports whether buildPosMap is the identity for s under the
// given lowercasing mode, i.e. whether posMap[i] is i for every index the
// WordPiece split path can read. vocab records the answer once per token.
//
// The walk mirrors buildPosMap exactly, which is what makes a true result safe
// to act on: the two functions chunk s the same way, so they always agree on
// the codepoint boundaries. Without lowercasing the identity only breaks on a
// control byte or a dropped codepoint; with it, a codepoint whose lowercase NFD
// form contributes a number of entries other than one breaks it as well.
func posMapIdentity(s string, doLowerCase bool) bool {
	for cur := 0; cur < len(s); {
		c := s[cur]
		if c < 0x80 {
			if isCntrlByte(c) {
				return false
			}
			cur++
			continue
		}

		n := charLen(s[cur:])
		chunk := s[cur : cur+n]
		r := decodeChar(chunk)
		cur += n

		if classifyRune(r) == classDrop {
			return false
		}
		if doLowerCase {
			lower := toLowerRune(r)
			// A codepoint that has no lowercase mapping of its own, is not a
			// nonspacing mark and is already in NFD form contributes exactly
			// one entry, which is the answer for most of a vocabulary.
			// QuickSpanString only reports a prefix it verified, so a short
			// answer falls through to the exact count.
			if lower != r || isMn(r) || norm.NFD.QuickSpanString(chunk) != len(chunk) {
				if nfdCodepointNumber(encodeRune(lower)) != 1 {
					return false
				}
			}
		}
	}
	return true
}
