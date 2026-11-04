package integer

import (
	"errors"
	"simd"
)

var ErrCarrylessWidth = errors.New("unsupported carryless Uint64 width")

// Clmul64Scalar multiplies polynomials over GF(2) modulo x^64. All lengths are
// checked before writes. Exact aliases succeed; other overlap is unsupported.
func Clmul64Scalar(a, b, out []uint64) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i, x := range a {
		y := b[i]
		var product uint64
		for y != 0 {
			if y&1 != 0 {
				product ^= x
			}
			x <<= 1
			y >>= 1
		}
		out[i] = product
	}
	return nil
}

// Clmul64SIMD retains low words of both even/odd 128-bit SDK products. Scratch
// stores, extraction and output writes are deliberately inside the timed call.
func Clmul64SIMD(a, b, out []uint64) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	var v simd.Uint64s
	w := v.Len()
	if w != 2 && w != 4 && w != 8 {
		return ErrCarrylessWidth
	}
	for i := 0; i < len(a); {
		end := i + min(w, len(a)-i)
		x, n := simd.LoadUint64sPart(a[i:end])
		y, _ := simd.LoadUint64sPart(b[i:end])
		var even, odd [8]uint64
		x.CarrylessMultiplyEven(y).StorePart(even[:w])
		x.CarrylessMultiplyOdd(y).StorePart(odd[:w])
		for j := 0; j < n; j++ {
			if j&1 == 0 {
				out[i+j] = even[j]
			} else {
				out[i+j] = odd[j-1]
			}
		}
		i += n
	}
	return nil
}
