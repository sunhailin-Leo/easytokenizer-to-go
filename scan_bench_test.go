package tokenizer

import (
	"strings"
	"testing"
)

// The scan primitives only take their vector path once the input is at least
// one vector wide, so the benchmarks below cover both sides of that boundary.

var (
	benchAlnum  = strings.Repeat("aZ9", 40) + "!"
	benchAlnumS = strings.Repeat("aZ9", 4) + "!"
	benchCJK    = strings.Repeat("广东省深圳市南山区腾讯滨海大厦", 3)
	benchPlain  = strings.Repeat("hello world, natural language processing. ", 6)
	benchShort  = "natural"
)

func BenchmarkScanAlnumRun(b *testing.B) {
	for b.Loop() {
		alnumRun(benchAlnum)
	}
}

func BenchmarkScanAlnumRunShort(b *testing.B) {
	for b.Loop() {
		alnumRun(benchAlnumS)
	}
}

func BenchmarkScanAllASCIIAlnum(b *testing.B) {
	for b.Loop() {
		allASCIIAlnum(benchPlain)
	}
}

func BenchmarkScanCountCodepointsASCII(b *testing.B) {
	for b.Loop() {
		countCodepoints(benchPlain)
	}
}

func BenchmarkScanCountCodepointsCJK(b *testing.B) {
	for b.Loop() {
		countCodepoints(benchCJK)
	}
}

func BenchmarkScanCountCodepointsShort(b *testing.B) {
	for b.Loop() {
		countCodepoints(benchShort)
	}
}

func BenchmarkScanNextSpecialStart(b *testing.B) {
	for b.Loop() {
		nextSpecialStart(benchPlain, '[')
	}
}

// nextSpecialStart delegates to strings.IndexByte, whose arm64 kernel is
// NEON assembly. This scalar loop is what it replaced; it is kept here so the
// comparison can be re-measured on another machine.
func BenchmarkScanNextSpecialStartScalar(b *testing.B) {
	for b.Loop() {
		scanNextSpecialStart(benchPlain)
	}
}

func BenchmarkScanNextSpecialStartHit(b *testing.B) {
	in := benchPlain + " [SEP]"
	for b.Loop() {
		nextSpecialStart(in, '[')
	}
}

func BenchmarkScanNextSpecialStartHitScalar(b *testing.B) {
	in := benchPlain + " [SEP]"
	for b.Loop() {
		scanNextSpecialStart(in)
	}
}

func BenchmarkScanNextSpecialStartShort(b *testing.B) {
	for b.Loop() {
		nextSpecialStart("natural", '[')
	}
}

func BenchmarkScanNextSpecialStartShortScalar(b *testing.B) {
	for b.Loop() {
		scanNextSpecialStart("natural")
	}
}
