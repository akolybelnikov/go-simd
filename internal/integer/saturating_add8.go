// Package integer implements explicitly typed integer laboratory kernels.
package integer

import (
	"errors"
	"simd"
)

var ErrLengthMismatch = errors.New("integer input/output slice lengths must match")

// SaturatingAdd8Scalar returns min(a+b,255), never modulo-256 wrapping.
// Length errors precede writes; exact aliases are supported, partial overlap not.
func SaturatingAdd8Scalar(a, b, out []uint8) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := range a {
		s := uint16(a[i]) + uint16(b[i])
		if s > 255 {
			s = 255
		}
		out[i] = uint8(s)
	}
	return nil
}
func SaturatingAdd8SIMD(a, b, out []uint8) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	var vector simd.Uint8s
	w := vector.Len()
	for i := 0; i < len(a); {
		end := i + min(w, len(a)-i)
		x, n := simd.LoadUint8sPart(a[i:end])
		y, _ := simd.LoadUint8sPart(b[i:end])
		x.AddSaturated(y).StorePart(out[i:end])
		i += n
	}
	return nil
}
