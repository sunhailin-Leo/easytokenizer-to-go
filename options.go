package tokenizer

import (
	"io"
	"io/fs"
)

// OffsetType is the unit of the offsets reported by [EasyTokenizer.EncodeWithIds],
// [EasyTokenizer.WordPieceTokenize] and the [Encoding] methods.
type OffsetType int

const (
	// OffsetCodepoint counts UTF-8 codepoints. It is what this tokenizer
	// reports: the original C++ implementation measured offsets this way, and
	// [ByteOffset] converts one to a byte offset for slicing.
	OffsetCodepoint OffsetType = iota
	// OffsetByte counts bytes. Nothing in this package reports byte offsets
	// yet; the constant exists so callers can branch on the unit instead of
	// hard-coding it.
	OffsetByte
)

// SpecialTokens names the tokens the tokenizer keeps whole: they are never
// split, and their ids are resolved once at construction.
//
// A field set to the empty string removes that token. It is then not kept
// whole and its id resolves to the UNK fallback, which is what a vocabulary
// without that token needs. SpecialTokens{} therefore removes all five;
// start from [DefaultSpecialTokens] instead.
type SpecialTokens struct {
	Pad  string
	CLS  string
	SEP  string
	UNK  string
	Mask string
}

// DefaultSpecialTokens returns the five special tokens BERT vocabularies use.
func DefaultSpecialTokens() SpecialTokens {
	return SpecialTokens{
		Pad:  padToken,
		CLS:  clsToken,
		SEP:  sepToken,
		UNK:  unkToken,
		Mask: maskToken,
	}
}

// Option configures a tokenizer at construction. The constructors that return
// an error accept options; [NewTokenizer] uses the defaults.
type Option func(*config)

type config struct {
	special SpecialTokens
}

// WithSpecialTokens replaces the special tokens. Every token is matched as it
// is spelled in the vocabulary file.
func WithSpecialTokens(tokens SpecialTokens) Option {
	return func(c *config) { c.special = tokens }
}

// NewTokenizerWithOptions creates a tokenizer from a vocabulary file, which
// must contain one token per line.
//
// The options are the same as for the other error returning constructors; use
// [WithSpecialTokens] for a vocabulary that spells the special tokens
// differently.
func NewTokenizerWithOptions(vocabPath string, doLowerCase bool, opts ...Option) (*EasyTokenizer, error) {
	v, err := loadVocab(vocabPath, doLowerCase)
	if err != nil {
		return nil, err
	}
	return newTokenizer(v, doLowerCase, opts), nil
}

// NewTokenizerFromReader creates a tokenizer from a vocabulary read from r,
// which must contain one token per line.
func NewTokenizerFromReader(r io.Reader, doLowerCase bool, opts ...Option) (*EasyTokenizer, error) {
	v, err := loadVocabReader(r, doLowerCase)
	if err != nil {
		return nil, err
	}
	return newTokenizer(v, doLowerCase, opts), nil
}

// NewTokenizerFromFS creates a tokenizer from a vocabulary file inside fsys.
// It accepts an [embed.FS], so a vocabulary embedded with go:embed needs no
// temporary file:
//
//	//go:embed vocab.txt
//	var vocabFS embed.FS
//
//	tk, err := tokenizer.NewTokenizerFromFS(vocabFS, "vocab.txt", true)
//
// [os.DirFS] and [testing/fstest.MapFS] work as well.
func NewTokenizerFromFS(fsys fs.FS, path string, doLowerCase bool, opts ...Option) (*EasyTokenizer, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}
	return newTokenizer(loadVocabBytes(data, doLowerCase), doLowerCase, opts), nil
}

// newTokenizer assembles a tokenizer from a loaded vocabulary.
func newTokenizer(v *vocab, doLowerCase bool, opts []Option) *EasyTokenizer {
	cfg := config{special: DefaultSpecialTokens()}
	for _, opt := range opts {
		opt(&cfg)
	}

	t := &EasyTokenizer{
		vocab:          v,
		basic:          newBasicTokenizer(doLowerCase, cfg.special),
		codePointLevel: true,
		special:        cfg.special,
	}
	t.padID = v.idOrUnknown(cfg.special.Pad)
	t.clsID = v.idOrUnknown(cfg.special.CLS)
	t.sepID = v.idOrUnknown(cfg.special.SEP)
	t.unkID = v.idOrUnknown(cfg.special.UNK)
	t.maskID = v.idOrUnknown(cfg.special.Mask)
	t.specialIDs = make([]int32, len(t.basic.special.keys))
	for i, key := range t.basic.special.keys {
		t.specialIDs[i] = v.id(key, t.unkID)
	}
	return t
}

// specialScan is the byte level scan that finds the first byte of a special
// token in the input. Almost every vocabulary spells all five special tokens
// with a leading '[' or '<', so the scan is a single [strings.IndexByte] over
// that byte in the common case and a scalar scan over the set of leading bytes
// otherwise.
type specialScan struct {
	first [256]bool
	lead  byte
	same  bool
	any   bool
}

func newSpecialScan(keys []string) specialScan {
	var s specialScan
	if len(keys) == 0 {
		return s
	}
	s.any = true
	s.same = true
	s.lead = keys[0][0]
	for _, key := range keys {
		s.first[key[0]] = true
		if key[0] != s.lead {
			s.same = false
		}
	}
	return s
}

// next returns the offset of the next byte that can start a special token, or
// -1 when the rest of the input holds none.
func (s *specialScan) next(text string) int {
	if !s.any {
		return -1
	}
	if s.same {
		return nextSpecialStart(text, s.lead)
	}
	for i := 0; i < len(text); i++ {
		if s.first[text[i]] {
			return i
		}
	}
	return -1
}
