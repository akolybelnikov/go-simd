package movement

import "simd"

// Copy64Scalar moves exact binary64 representations without arithmetic.
// Equal lengths are checked before writes; exact aliases are supported,
// partial overlap is outside this experiment's common contract.
func Copy64Scalar(x, out []float64) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i := range x {
		out[i] = x[i]
	}
	return nil
}
func Copy64Go(x, out []float64) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	copy(out, x)
	return nil
}
func Copy64SIMD(x, out []float64) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	var v simd.Float64s
	w := v.Len()
	for i := 0; i < len(x); {
		end := i + min(w, len(x)-i)
		a, n := simd.LoadFloat64sPart(x[i:end])
		a.StorePart(out[i:end])
		i += n
	}
	return nil
}
