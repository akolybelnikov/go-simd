package addition

import "simd"

// Add64Scalar adds true binary64 operands. All length errors precede writes;
// disjoint/exact aliases are supported, partial output overlap is unsupported.
func Add64Scalar(a, b, out []float64) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return nil
}

// Add64SIMD includes partial loads/stores at the actual float64 lane width.
func Add64SIMD(a, b, out []float64) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	var vector simd.Float64s
	width := vector.Len()
	for i := 0; i < len(a); {
		// The selected SDK's emulated partial load reports the slice length.
		// Bound it to real lanes so no elements are skipped in that mode.
		end := i + min(width, len(a)-i)
		x, n := simd.LoadFloat64sPart(a[i:end])
		y, _ := simd.LoadFloat64sPart(b[i:end])
		x.Add(y).StorePart(out[i:end])
		i += n
	}
	return nil
}
