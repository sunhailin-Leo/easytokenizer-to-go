//go:build goexperiment.simd

package tokenizer

import "simd"

// Vectorised counterparts of the scanning primitives in scan_nosimd.go, used
// when the module is built with GOEXPERIMENT=simd. Both builds must agree on
// every input.
//
// The portable simd package sizes its vectors at run time (simd.VectorBitSize)
// and has no cross lane reduction, so a kernel that needs to know where a mask
// changes has to materialise it into a stack buffer and scan the lanes. Inputs
// too short to fill a vector go to the scalar version, which is also what a
// build without the experiment always uses.

// maxVectorBytes is the widest vector the simd package reports, i.e. the 64
// bytes of AVX-512. Kernels fall back to scalar code on anything wider.
const maxVectorBytes = 64

// vectorBytes is the vector width of the running program.
var vectorBytes = simd.VectorBitSize() / 8

func alnumRun(s string) (int, bool) {
	if vectorBytes <= maxVectorBytes {
		// Most runs end inside the first vector, so scan that much scalar and
		// only pay for the vector setup when the run clearly continues.
		n, hasUpper := scalarAlnumRunBudget(s, vectorBytes)
		if n < vectorBytes {
			return n, hasUpper
		}
		return alnumRunVector(s, n, hasUpper)
	}
	return scalarAlnumRun(s)
}

func countCodepoints(s string) int {
	if vectorBytes <= maxVectorBytes && len(s) >= vectorBytes {
		return countCodepointsVector(s)
	}
	return scalarCountCodepoints(s)
}

// asciiAlnumMask returns the masks of the ASCII alphanumeric bytes of v and of
// its uppercase bytes.
//
// Uint8s has no Greater or Less, so a range test is written as
// min(x-base, span) == x-base, which is true exactly when x-base <= span.
func asciiAlnumMask(v simd.Uint8s) (alnum, upper simd.Mask8s) {
	letters := v.Or(simd.BroadcastUint8s(0x20)).Sub(simd.BroadcastUint8s('a'))
	letter := letters.Min(simd.BroadcastUint8s('z' - 'a')).Equal(letters)
	digits := v.Sub(simd.BroadcastUint8s('0'))
	digit := digits.Min(simd.BroadcastUint8s(9)).Equal(digits)
	capital := v.Sub(simd.BroadcastUint8s('A'))
	upper = capital.Min(simd.BroadcastUint8s('Z' - 'A')).Equal(capital)
	return letter.Or(digit), upper
}

// storeLanes writes the 0xff/0x00 lanes of mask into buf, which must hold at
// least vectorBytes bytes.
func storeLanes(mask simd.Mask8s, buf *[maxVectorBytes]uint8) {
	mask.ToInt8s().ToBits().Store(buf[:vectorBytes])
}

// loadVector loads the vectorBytes bytes of s starting at i.
func loadVector(s string, i int) simd.Uint8s {
	return simd.LoadUint8s([]byte(s[i : i+vectorBytes]))
}

// firstLaneDiffersFrom returns the index of the first lane that is not want,
// or -1 when every lane matches.
func firstLaneDiffersFrom(lanes []uint8, want uint8) int {
	for i, v := range lanes {
		if v != want {
			return i
		}
	}
	return -1
}

// anyLaneSet reports whether any lane is non zero. It stores v into buf, which
// it may overwrite.
func anyLaneSet(v simd.Uint8s, buf *[maxVectorBytes]uint8) bool {
	v.Store(buf[:vectorBytes])
	for _, x := range buf[:vectorBytes] {
		if x != 0 {
			return true
		}
	}
	return false
}

// alnumRunVector continues the run that starts with n already scanned bytes,
// none of which was a non alphanumeric byte.
func alnumRunVector(s string, n int, hasUpper bool) (int, bool) {
	var lanes [maxVectorBytes]uint8
	var upper simd.Uint8s

	i := n
	for len(s)-i >= vectorBytes {
		alnum, up := asciiAlnumMask(loadVector(s, i))
		storeLanes(alnum, &lanes)
		if firstLaneDiffersFrom(lanes[:vectorBytes], 0xFF) >= 0 {
			// The run ends inside this vector, so its uppercase letters are
			// left to the scalar scan below together with the ones after it.
			break
		}
		upper = upper.Or(up.ToInt8s().ToBits())
		i += vectorBytes
	}

	if i > n && anyLaneSet(upper, &lanes) {
		hasUpper = true
	}
	tail, tailUpper := scalarAlnumRun(s[i:])
	return i + tail, hasUpper || tailUpper
}

// countCodepointsVector counts the bytes that start a UTF-8 codepoint, which
// are the bytes whose top two bits are not 10. Each lane accumulates at most
// one per iteration, so the accumulator is flushed well before it can wrap.
func countCodepointsVector(s string) int {
	cont := simd.BroadcastUint8s(0xC0)
	want := simd.BroadcastUint8s(0x80)
	ones := simd.BroadcastUint8s(1)

	var lanes [maxVectorBytes]uint8
	acc := simd.Uint8s{}
	continuations := 0

	i, flushed := 0, 0
	for len(s)-i >= vectorBytes {
		v := loadVector(s, i)
		acc = acc.Add(ones.Masked(v.And(cont).Equal(want)))
		i += vectorBytes
		if i-flushed >= 255*vectorBytes {
			acc.Store(lanes[:vectorBytes])
			for _, x := range lanes[:vectorBytes] {
				continuations += int(x)
			}
			acc = simd.Uint8s{}
			flushed = i
		}
	}
	acc.Store(lanes[:vectorBytes])
	for _, x := range lanes[:vectorBytes] {
		continuations += int(x)
	}

	for ; i < len(s); i++ {
		if s[i]&0xC0 == 0x80 {
			continuations++
		}
	}
	return len(s) - continuations
}
