package patterns

import (
	"math"
	"simd"
)

// BitRoundTripScalar preserves the exact opaque representation, including NaNs.
// Length errors precede writes; empty/exact aliases work, partial overlap not.
func BitRoundTripScalar(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		out[i] = math.Float32frombits(math.Float32bits(v))
	}
	return nil
}

// BitRoundTripSIMD includes the reinterpretation path in the measured kernel.
// Its source does not establish which bitcast instructions the compiler emits.
func BitRoundTripSIMD(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	var vector simd.Float32s
	w := vector.Len()
	for i := 0; i < len(x); {
		end := i + min(w, len(x)-i)
		v, n := simd.LoadFloat32sPart(x[i:end])
		v.ToBits().BitsToFloat32().StorePart(out[i:end])
		i += n
	}
	return nil
}
