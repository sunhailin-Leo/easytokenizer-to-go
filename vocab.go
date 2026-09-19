package tokenizer

import (
	"io"
	"os"
	"strings"
)

// The special tokens the tokenizer knows about, and their default text. They
// are looked up in the vocabulary file to find their ids.
const (
	padToken  = "[PAD]"
	clsToken  = "[CLS]"
	sepToken  = "[SEP]"
	unkToken  = "[UNK]"
	maskToken = "[MASK]"
)

// vocab stores the token to id mapping of a vocabulary file.
//
// ids is used for the exact lookups that dominate WordPiece splitting, while
// trie answers the longest-prefix queries the WordPiece loop needs and keeps
// the tokens in id order.
type vocab struct {
	ids    map[string]int32
	tokens []string
	trie   *byteTrie
	// posMapIdentity[i] reports whether buildPosMap is the identity for
	// tokens[i], which lets the WordPiece split path skip building the map for
	// the words that do not need it. It is fixed at construction because a live
	// tokenizer is read-only.
	posMapIdentity []bool
}

func loadVocab(path string, doLowerCase bool) (*vocab, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loadVocabBytes(data, doLowerCase), nil
}

// loadVocabReader reads a vocabulary from r, which must contain one token per
// line, and parses it.
func loadVocabReader(r io.Reader, doLowerCase bool) (*vocab, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return loadVocabBytes(data, doLowerCase), nil
}

// loadVocabBytes parses a vocabulary buffer.
func loadVocabBytes(data []byte, doLowerCase bool) *vocab {
	s := string(data)

	// One token per line, so the newline count sizes the map and the token
	// slice without a second pass.
	lines := strings.Count(s, "\n") + 1
	v := &vocab{
		ids:    make(map[string]int32, lines),
		tokens: make([]string, 0, lines),
	}
	// Lines are split the way std::getline did: on '\n', keeping a trailing
	// '\r'. Empty lines are skipped.
	for len(s) > 0 {
		line := s
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			line, s = s[:i], s[i+1:]
		} else {
			s = ""
		}
		if line == "" {
			continue
		}
		// Duplicated tokens keep the id they were first seen with, which is
		// what the double array trie of the C++ implementation did.
		if _, ok := v.ids[line]; ok {
			continue
		}
		v.ids[line] = int32(len(v.tokens))
		v.tokens = append(v.tokens, line)
	}

	v.trie = buildTrie(v.tokens)
	v.posMapIdentity = make([]bool, len(v.tokens))
	for i, token := range v.tokens {
		v.posMapIdentity[i] = posMapIdentity(token, doLowerCase)
	}
	return v
}

// lookup returns the id of the token and whether the vocabulary holds it.
func (v *vocab) lookup(token string) (int32, bool) {
	id, ok := v.ids[token]
	return id, ok
}

// posMapIsIdentity reports whether buildPosMap is the identity for the token
// with this id.
func (v *vocab) posMapIsIdentity(id int32) bool {
	return v.posMapIdentity[id]
}

// id returns the id of the token, or fallback when it is not in the
// vocabulary.
func (v *vocab) id(token string, fallback int32) int32 {
	if id, ok := v.ids[token]; ok {
		return id
	}
	return fallback
}

// idOrUnknown returns the id of the token, or -1 when it is missing from the
// vocabulary. The C++ implementation stored the missing special tokens as
// SizeT(-1), which surfaces as -1 in the int32 id sequences.
func (v *vocab) idOrUnknown(token string) int32 {
	if id, ok := v.ids[token]; ok {
		return id
	}
	return -1
}
