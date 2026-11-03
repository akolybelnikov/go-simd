package patterns

import (
	"simd"

	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
)

// FMAPath exposes the shared cached runtime capability gate.
func FMAPath() string { return fmacap.Path() }

// AffineFMAScalar rounds alpha*x+beta once to binary32, including subnormals.
func AffineFMAScalar(x, out []float32, alpha, beta float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		out[i] = fma32.FMA(alpha, v, beta)
	}
	return nil
}

// AffineFMASIMD uses verified vector MulAdd, or the explicitly recorded software
// path. The process gate is outside timing; validation/dispatch/broadcasts remain.
func AffineFMASIMD(x, out []float32, alpha, beta float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	if !fmacap.VectorAvailable() {
		for i, v := range x {
			out[i] = fma32.FMA(alpha, v, beta)
		}
		return nil
	}
	a, b := simd.BroadcastFloat32s(alpha), simd.BroadcastFloat32s(beta)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		a.MulAdd(v, b).StorePart(out[i:])
		i += n
	}
	return nil
}
