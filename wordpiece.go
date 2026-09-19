package tokenizer

// maxInputCharsPerWord bounds the byte length of a word the WordPiece
// algorithm is willing to split; longer words become a single UNK token.
const maxInputCharsPerWord = 100

// subToken is a WordPiece piece of a word, with positions measured in
// codepoints relative to the start of the word. id is the vocabulary id of the
// piece, which is also the id of the token the piece becomes.
type subToken struct {
	start int32
	end   int32
	text  string
	id    int32
}

// wpKind selects the half of the WordPiece result a caller needs.
type wpKind uint8

const (
	wpTokenTexts wpKind = iota // the token strings
	wpTokenIds                 // the id of every token
	wpTokenAll                 // the token strings and their ids
)

// wpWalk collects the result of one WordPiece walk: the token strings or their
// ids, plus the codepoint level offsets of every token.
//
// The ids are resolved while the walk runs, so no caller has to look a token up
// again, and a walk that only feeds an id sequence never materialises the token
// strings.
type wpWalk struct {
	kind    wpKind
	texts   []string
	ids     []int32
	offsets []int32
}

// add records one token: its text, its id, or both, depending on the kind of
// the walk.
func (w *wpWalk) add(text string, id int32) {
	switch w.kind {
	case wpTokenIds:
		w.ids = append(w.ids, id)
	case wpTokenTexts:
		w.texts = append(w.texts, text)
	default:
		w.ids = append(w.ids, id)
		w.texts = append(w.texts, text)
	}
}

// wordpieceWalk splits text into WordPiece tokens.
func (t *EasyTokenizer) wordpieceWalk(text string, kind wpKind) wpWalk {
	baseTokens := t.basic.basicTokenize(text, nil)

	w := wpWalk{kind: kind}
	// Every base token produces at least one piece, so the base token count is
	// a lower bound for the result slices.
	switch kind {
	case wpTokenIds:
		w.ids = make([]int32, 0, len(baseTokens))
	case wpTokenTexts:
		w.texts = make([]string, 0, len(baseTokens))
	default:
		w.ids = make([]int32, 0, len(baseTokens))
		w.texts = make([]string, 0, len(baseTokens))
	}
	w.offsets = make([]int32, 0, 2*len(baseTokens))

	var byteToIndex []int32
	if t.codePointLevel {
		byteToIndex = buildIndexMap(text)
	}

	var posMap []int32
	var subTokens []subToken

	for _, base := range baseTokens {
		start, end := int(base.start), int(base.end)
		word := tokenText(text, base)

		// The vocabulary is asked first: a word that is part of it, which is
		// the common case, then skips the special token trie walk.
		if id, ok := t.vocab.lookup(word); ok {
			w.add(word, id)
			w.offsets = t.appendOffsets(w.offsets, byteToIndex, start, end)
			continue
		}
		if k, special := t.basic.special.exact(word); special {
			// A special token that is missing from the vocabulary falls back to
			// the UNK id, like every other unknown token.
			w.add(word, t.specialIDs[k])
			w.offsets = t.appendOffsets(w.offsets, byteToIndex, start, end)
			continue
		}

		if len(word) > maxInputCharsPerWord {
			w.add(unkToken, t.unkID)
			w.offsets = t.appendOffsets(w.offsets, byteToIndex, start, end)
			continue
		}

		// Split the word into the longest vocabulary pieces starting at the
		// current position, prefixing continuation pieces with "##".
		subTokens = subTokens[:0]
		cur, pos := 0, int32(0)
		isBad := false
		// identity records whether buildPosMap maps every index to itself for
		// the word, judged from its pieces, which holds for almost every word
		// and lets the offset loop below use the positions it already has.
		identity := true
		for cur < len(word) {
			var id int32
			var ok bool
			if cur == 0 {
				id, ok = t.vocab.trie.longestPrefix(word)
			} else {
				id, ok = t.vocab.trie.longestPrefixParts("##", word[cur:])
			}
			if !ok || (cur > 0 && len(t.vocab.tokens[id]) < 3) {
				isBad = true
				break
			}
			prefix := t.vocab.tokens[id]

			n := len(prefix)
			chars := int32(getCodepointNumber(prefix))
			if cur > 0 {
				n -= 2
				chars -= 2
			}
			subTokens = append(subTokens, subToken{start: pos, end: pos + chars, text: prefix, id: id})
			if !t.vocab.posMapIsIdentity(id) {
				identity = false
			}
			cur += n
			pos += chars
		}

		if isBad {
			w.add(unkToken, t.unkID)
			w.offsets = t.appendOffsets(w.offsets, byteToIndex, start, end)
			continue
		}

		if allASCIIAlnum(word) {
			for _, sub := range subTokens {
				w.add(sub.text, sub.id)
				if t.codePointLevel {
					base := byteToIndex[start]
					w.offsets = append(w.offsets, base+sub.start, base+sub.end)
				} else {
					w.offsets = append(w.offsets, int32(start)+sub.start, int32(start)+sub.end)
				}
			}
			continue
		}

		// The flags are recorded per vocabulary token and the pieces tile the
		// word, so they settle the word as well; a boundary inside a codepoint
		// makes the piece after it start with a continuation byte, which is a
		// dropped codepoint and clears the flags.
		if !identity {
			posMap = posMap[:0]
			posMap = buildPosMap(word, t.basic.doLowerCase, posMap)
		}
		for _, sub := range subTokens {
			a, b := sub.start, sub.end
			if !identity {
				if posMap[a] == posMap[b] {
					b = posMap[a] + 1
				} else {
					b = posMap[b]
				}
				a = posMap[a]
			}

			w.add(sub.text, sub.id)
			if t.codePointLevel {
				base := byteToIndex[start]
				w.offsets = append(w.offsets, base+a, base+b)
			} else {
				w.offsets = append(w.offsets,
					int32(start+search(word, int(a))),
					int32(start+search(word, int(b))))
			}
		}
	}

	return w
}

// appendOffsets appends the offset pair of a token that was not split
// further.
func (t *EasyTokenizer) appendOffsets(offsets []int32, byteToIndex []int32, start, end int) []int32 {
	if t.codePointLevel {
		return append(offsets, byteToIndex[start], byteToIndex[end])
	}
	return append(offsets, int32(start), int32(end))
}
