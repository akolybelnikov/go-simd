package patterns

import (
	"math"
	"simd"
)

// XorBitsScalar operates on opaque representations, with no floating arithmetic.
func XorBitsScalar(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := range a {
		out[i] = math.Float32frombits(math.Float32bits(a[i]) ^ math.Float32bits(b[i]))
	}
	return nil
}

// XorBitsSIMD supports exact aliases; partial output overlap is unsupported.
func XorBitsSIMD(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	var vector simd.Float32s
	w := vector.Len()
	for i := 0; i < len(a); {
		end := i + min(w, len(a)-i)
		x, n := simd.LoadFloat32sPart(a[i:end])
		y, _ := simd.LoadFloat32sPart(b[i:end])
		x.ToBits().Xor(y.ToBits()).BitsToFloat32().StorePart(out[i:end])
		i += n
	}
	return nil
}
