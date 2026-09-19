//go:build !goexperiment.simd

package tokenizer

// Scalar implementations of the scanning primitives, used when the simd
// experiment is not enabled. Builds with GOEXPERIMENT=simd use the vectorised
// versions in scan_simd.go instead; both must behave identically.

// alnumRun returns the number of leading ASCII alphanumeric bytes of s and
// whether any of them is an uppercase letter.
func alnumRun(s string) (int, bool) {
	return scalarAlnumRun(s)
}

// countCodepoints returns the number of UTF-8 encoded codepoints in s.
func countCodepoints(s string) int {
	return scalarCountCodepoints(s)
}
