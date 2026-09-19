package tokenizer

import "testing"

func TestByteTrieExactAndPrefix(t *testing.T) {
	keys := []string{"[PAD]", "[CLS]", "a", "ab", "abc", "abd", "b", "น", "นค", "##a", "##ab"}
	tr := buildTrie(keys)

	for i, key := range keys {
		got, ok := tr.exact(key)
		if !ok {
			t.Fatalf("exact(%q): not found", key)
		}
		if int(got) != i {
			t.Fatalf("exact(%q) = %d, want %d", key, got, i)
		}
	}
	if _, ok := tr.exact("abcd"); ok {
		t.Fatal("exact(abcd) should not be found")
	}

	cases := []struct {
		in   string
		want string
	}{
		{"abc", "abc"},
		{"abd", "abd"},
		{"abcd", "abc"},
		{"ab", "ab"},
		{"a", "a"},
		{"b", "b"},
		{"นครปฐม", "นค"},
		{"น", "น"},
		{"##abc", "##ab"},
		{"zzz", ""},
	}
	for _, c := range cases {
		id, ok := tr.longestPrefix(c.in)
		got := ""
		if ok {
			got = keys[id]
		}
		if got != c.want {
			t.Errorf("longestPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestByteTrieLongestPrefixParts(t *testing.T) {
	keys := []string{"#", "##", "##a", "##ab", "##abc", "a", "ab", "abd"}
	tr := buildTrie(keys)

	for _, suffix := range []string{"", "a", "ab", "abc", "abcd", "abd", "b", "z"} {
		want, wantOK := tr.longestPrefix("##" + suffix)
		got, gotOK := tr.longestPrefixParts("##", suffix)
		if gotOK != wantOK || (wantOK && got != want) {
			t.Errorf("longestPrefixParts(%q, %q) = (%d, %v), want (%d, %v)",
				"##", suffix, got, gotOK, want, wantOK)
		}
	}

	if _, ok := tr.longestPrefixParts("xy", "abc"); ok {
		t.Error("longestPrefixParts with an unknown prefix should not match")
	}
}

func TestBasicTokenizerSpecialTokens(t *testing.T) {
	b := newBasicTokenizer(true, DefaultSpecialTokens())
	text := "[CLS]x[UNK]"
	matches := b.specialTokens(text)

	want := []specialMatch{{pos: 0, text: "[CLS]"}, {pos: 6, text: "[UNK]"}}
	if len(matches) != len(want) {
		t.Fatalf("specialTokens(%q) = %v, want %v", text, matches, want)
	}
	for i := range want {
		if matches[i] != want[i] {
			t.Fatalf("specialTokens(%q) = %v, want %v", text, matches, want)
		}
	}
}
