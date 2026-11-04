package patterns

import (
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"simd"
)

// MulAddScalar rounds a*b+c once to binary32. NaN payloads are unspecified.
// Length errors precede writes. Disjoint and exact input aliases are supported;
// partial output/input overlap is outside the contract.
func MulAddScalar(a, b, c, out []float32) error {
	if len(a) != len(b) || len(a) != len(c) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := range a {
		out[i] = fma32.FMA(a[i], b[i], c[i])
	}
	return nil
}

// MulAddSIMD uses verified portable vector FMA or the recorded software path.
// Validation and dispatch are per call; no broadcast or reduction is performed.
func MulAddSIMD(a, b, c, out []float32) error {
	if len(a) != len(b) || len(a) != len(c) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	if !fmacap.VectorAvailable() {
		for i := range a {
			out[i] = fma32.FMA(a[i], b[i], c[i])
		}
		return nil
	}
	for i := 0; i < len(a); {
		x, n := simd.LoadFloat32sPart(a[i:])
		y, _ := simd.LoadFloat32sPart(b[i:])
		z, _ := simd.LoadFloat32sPart(c[i:])
		x.MulAdd(y, z).StorePart(out[i:])
		i += n
	}
	return nil
}
