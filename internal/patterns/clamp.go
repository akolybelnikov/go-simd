package patterns

import (
	"errors"
	"simd"
)

var ErrInvalidBounds = errors.New("clamp bounds must be ordered and non-NaN")

// ClampScalar clips only ordered values outside [lower,upper]. Equality and
// unordered input preserve the original bits, including signed zeros/NaNs.
// Length errors take precedence; invalid bounds fail before writes, even empty.
func ClampScalar(x, out []float32, lower, upper float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	if !(lower <= upper) {
		return ErrInvalidBounds
	}
	for i, v := range x {
		if v < lower {
			out[i] = lower
		} else if v > upper {
			out[i] = upper
		} else {
			out[i] = v
		}
	}
	return nil
}

// ClampSIMD uses ordered comparisons and bit-preserving selection. Partial
// stores exclude inactive input lanes. Broadcasts and validation are per call.
func ClampSIMD(x, out []float32, lower, upper float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	if !(lower <= upper) {
		return ErrInvalidBounds
	}
	lo, hi := simd.BroadcastFloat32s(lower), simd.BroadcastFloat32s(upper)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		clipped := lo.IfElse(v.Less(lo), v)
		hi.IfElse(v.Greater(hi), clipped).StorePart(out[i:])
		i += n
	}
	return nil
}

// ClampSIMDMinMax includes semantic repairs in its timed Min/Max pipeline.
// Raw Min/Max may select other bits for unordered/tied values. Repair original
// input ties/NaNs and exact outside-bound selection, including +0/-0 bounds.
func ClampSIMDMinMax(x, out []float32, lower, upper float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	if !(lower <= upper) {
		return ErrInvalidBounds
	}
	lo, hi := simd.BroadcastFloat32s(lower), simd.BroadcastFloat32s(upper)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		clipped := v.Max(lo).Min(hi)
		clipped = lo.IfElse(v.Less(lo), clipped)
		clipped = hi.IfElse(v.Greater(hi), clipped)
		keep := v.Equal(lo).Or(v.Equal(hi)).Or(v.NotEqual(v))
		v.IfElse(keep, clipped).StorePart(out[i:])
		i += n
	}
	return nil
}
