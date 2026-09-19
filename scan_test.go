package tokenizer

import (
	"strings"
	"testing"
)

// kernelInputs holds strings that exercise the vector path and the scalar tail
// of every scanning kernel: long ASCII runs, runs that break on the first
// byte, mixed scripts, invalid UTF-8 and inputs shorter than a vector.
func kernelInputs() []string {
	inputs := []string{
		"",
		"a",
		"abcXYZ019",
		"abcXYZ019",
		strings.Repeat("a", 15),
		strings.Repeat("a", 16),
		strings.Repeat("a", 17),
		strings.Repeat("a", 31),
		strings.Repeat("a", 32),
		strings.Repeat("a", 33),
		strings.Repeat("a", 100),
		strings.Repeat("9", 64),
		strings.Repeat("Z", 64),
		strings.Repeat("z", 64),
		strings.Repeat("a", 63) + "_",
		strings.Repeat("a", 64) + "_",
		"广东省深圳市南山区腾讯滨海大厦",
		"นครปฐม เมืองนครปฐม ถนนขาด เลขที่ 69 หมู่ 1 ซ. - - ถ. -",
		"héllo wörld 日本語!",
		"[CLS]hello world[SEP]",
		strings.Repeat("x", 40) + "[UNK]" + strings.Repeat("y", 40),
		strings.Repeat("é", 40),
		strings.Repeat("\xc3", 40),
		strings.Repeat("\x80", 40),
		"\xff\xfe\xfd" + strings.Repeat("a", 40),
	}
	out := inputs[:0:0]
	return append(out, inputs...)
}

// scanAllASCIIAlnum is an independent reference implementation of the
// predicate every byte of s is an ASCII alphanumeric byte.
func scanAllASCIIAlnum(s string) bool {
	for _, c := range []byte(s) {
		switch {
		case '0' <= c && c <= '9', 'A' <= c && c <= 'Z', 'a' <= c && c <= 'z':
		default:
			return false
		}
	}
	return true
}

// scanNextSpecialStart is an independent reference implementation of the
// search for the next '['.
func scanNextSpecialStart(s string) int {
	for i := range len(s) {
		if s[i] == '[' {
			return i
		}
	}
	return -1
}

func TestScanKernelsMatchScalar(t *testing.T) {
	for _, in := range kernelInputs() {
		if gotN, gotUp := alnumRun(in); true {
			if wantN, wantUp := scalarAlnumRun(in); gotN != wantN || gotUp != wantUp {
				t.Errorf("alnumRun(%q) = %v, %v, want %v, %v", in, gotN, gotUp, wantN, wantUp)
			}
		}
		if got, want := allASCIIAlnum(in), scanAllASCIIAlnum(in); got != want {
			t.Errorf("allASCIIAlnum(%q) = %v, want %v", in, got, want)
		}
		if got, want := countCodepoints(in), scalarCountCodepoints(in); got != want {
			t.Errorf("countCodepoints(%q) = %d, want %d", in, got, want)
		}
		if got, want := nextSpecialStart(in, '['), scanNextSpecialStart(in); got != want {
			t.Errorf("nextSpecialStart(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestScanKernelsOnEveryOffset(t *testing.T) {
	base := "[CLS]广东省abcXYZ019นครปฐม 日本語 é[SEP]" + strings.Repeat("aZ9", 30)
	for i := 0; i <= len(base); i++ {
		in := base[i:]
		if gotN, gotUp := alnumRun(in); true {
			if wantN, wantUp := scalarAlnumRun(in); gotN != wantN || gotUp != wantUp {
				t.Fatalf("alnumRun(offset %d) = %v, %v, want %v, %v", i, gotN, gotUp, wantN, wantUp)
			}
		}
		if got, want := allASCIIAlnum(in), scanAllASCIIAlnum(in); got != want {
			t.Fatalf("allASCIIAlnum(offset %d) = %v, want %v", i, got, want)
		}
		if got, want := countCodepoints(in), scalarCountCodepoints(in); got != want {
			t.Fatalf("countCodepoints(offset %d) = %d, want %d", i, got, want)
		}
		if got, want := nextSpecialStart(in, '['), scanNextSpecialStart(in); got != want {
			t.Fatalf("nextSpecialStart(offset %d) = %d, want %d", i, got, want)
		}
	}
}
