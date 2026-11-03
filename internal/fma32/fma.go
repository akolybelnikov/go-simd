// Package fma32 implements one correctly rounded binary32 multiply-add.
// It has no dependency on SIMD or the experimental toolchain surface.
package fma32

import "math"

const SoftwarePath = "software-fma32/v1"
const VectorPath = "vector-fma32/v1"
const RoundingID = "ieee754-binary32-rne-once/v1"

// FMA returns a*b+c rounded once to float32, ties to even. NaN payloads are
// unspecified. Widening math.FMA and simply narrowing can double round.
func FMA(a, b, c float32) float32 {
	if math.Float32bits(a)&0x7f800000 == 0x7f800000 || math.Float32bits(b)&0x7f800000 == 0x7f800000 || math.Float32bits(c)&0x7f800000 == 0x7f800000 {
		return float32(math.FMA(float64(a), float64(b), float64(c)))
	}
	// A binary32 product is exact in binary64 (at most 48 significand bits).
	// Its full exponent range, including subnormal products, fits binary64.
	p := float64(float64(a) * float64(b))
	z := float64(c)
	s := float64(p + z)
	// Error-free TwoSum: exact a*b+c = s+residual. Explicit conversions keep
	// each intermediate rounded to binary64, independently of fusion choices.
	v := float64(s - p)
	residual := float64(float64(p-float64(s-v)) + float64(z-v))
	// Round to odd at binary64 precision before narrowing. If nearest-even s
	// is even and inexact, choose its odd neighbor toward the exact sum. With
	// 53 versus 24 bits this prevents midpoint double rounding, also at the
	// binary32 subnormal and overflow boundaries.
	if residual != 0 && math.Float64bits(s)&1 == 0 {
		direction := math.Inf(1)
		if residual < 0 {
			direction = math.Inf(-1)
		}
		s = math.Nextafter(s, direction)
	}
	return float32(s)
}
