// Package patterns explores broadcast constants and predicates on float32 arrays.
// Output may be disjoint or exactly alias any array input. Partial overlap is
// unsupported. Length errors never modify output. Broadcasts occur once per call.
package patterns

import (
	"errors"
	"simd"
)

var ErrLengthMismatch = errors.New("input and output slices must have equal lengths")

func ScaleScalar(x, out []float32, alpha float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i := range x {
		out[i] = alpha * x[i]
	}
	return nil
}

func ScaleSIMD(x, out []float32, alpha float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	a := simd.BroadcastFloat32s(alpha)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		v.Mul(a).StorePart(out[i:])
		i += n
	}
	return nil
}

// Affine rounds the product to float32 before adding beta, without fused FMA.
func AffineScalar(x, out []float32, alpha, beta float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i := range x {
		out[i] = float32(alpha*x[i]) + beta
	}
	return nil
}

func AffineSIMD(x, out []float32, alpha, beta float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	a, b := simd.BroadcastFloat32s(alpha), simd.BroadcastFloat32s(beta)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		v.Mul(a).Add(b).StorePart(out[i:])
		i += n
	}
	return nil
}

// ThresholdZero retains x only for ordered x > threshold; false yields +0.
func ThresholdZeroScalar(x, out []float32, threshold float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		if v > threshold {
			out[i] = v
		} else {
			out[i] = 0
		}
	}
	return nil
}

func ThresholdZeroSIMD(x, out []float32, threshold float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	t := simd.BroadcastFloat32s(threshold)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		v.Masked(v.Greater(t)).StorePart(out[i:])
		i += n
	}
	return nil
}

// SelectGT preserves the selected operand bits, including signed zero/NaN
// payloads. Equality or an unordered comparison takes y. Both arrays contribute
// to logical work accounting; this is not a claim about physical memory loads.
func SelectGTScalar(x, y, out []float32, threshold float32) error {
	if len(x) != len(y) || len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		if v > threshold {
			out[i] = v
		} else {
			out[i] = y[i]
		}
	}
	return nil
}

func SelectGTSIMD(x, y, out []float32, threshold float32) error {
	if len(x) != len(y) || len(x) != len(out) {
		return ErrLengthMismatch
	}
	t := simd.BroadcastFloat32s(threshold)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		other, _ := simd.LoadFloat32sPart(y[i:])
		v.IfElse(v.Greater(t), other).StorePart(out[i:])
		i += n
	}
	return nil
}

// RangeZero composes ordered strict bounds: lower < x && x < upper.
// Equality, unordered comparisons and empty/reversed ranges yield +0.
func RangeZeroScalar(x, out []float32, lower, upper float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		if v > lower && v < upper {
			out[i] = v
		} else {
			out[i] = 0
		}
	}
	return nil
}

func RangeZeroSIMD(x, out []float32, lower, upper float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	lo, hi := simd.BroadcastFloat32s(lower), simd.BroadcastFloat32s(upper)
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		v.Masked(v.Greater(lo).And(v.Less(hi))).StorePart(out[i:])
		i += n
	}
	return nil
}
