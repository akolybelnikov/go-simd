package patterns

import "simd"

// PredicateGTScalar exports ordered x > threshold as canonical uint8 0/1.
// Input/output must have equal lengths and disjoint storage. Invalid lengths
// leave output unchanged. Empty inputs succeed; NaN and equality produce 0.
func PredicateGTScalar(x []float32, out []uint8, threshold float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		out[i] = 0
		if v > threshold {
			out[i] = 1
		}
	}
	return nil
}

// PredicateGTSIMD compares partial vectors, selects explicit 0/1 values, then
// converts/extracts and writes valid bytes. Raw SDK mask encodings never escape.
// Broadcasts, scratch, conversion and writes belong to the measured call.
func PredicateGTSIMD(x []float32, out []uint8, threshold float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	t := simd.BroadcastFloat32s(threshold)
	one, zero := simd.BroadcastFloat32s(1), simd.BroadcastFloat32s(0)
	var scratch [16]int32
	lanes := scratch[:]
	if w := t.Len(); w > len(lanes) {
		lanes = make([]int32, w)
	}
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		one.IfElse(v.Greater(t), zero).ConvertToInt32().Store(lanes)
		for j := 0; j < n; j++ {
			out[i+j] = uint8(lanes[j])
		}
		i += n
	}
	return nil
}
