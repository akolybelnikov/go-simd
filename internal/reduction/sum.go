// Package reduction implements explicitly ordered float32 reductions.
package reduction

import "simd"

// SumScalar starts at +0 and rounds after each left-to-right addition.
func SumScalar(x []float32) float32 {
	var total float32
	for _, v := range x {
		total = float32(total + v)
	}
	return total
}

// SumSIMD adds full vectors into one striped accumulator, folds lanes from
// low to high starting at +0, then adds the scalar tail left to right. The
// extraction and final fold are part of the kernel. No padded lanes are read.
func SumSIMD(x []float32) float32 {
	if len(x) == 0 {
		return 0
	}
	var accumulator simd.Float32s
	w := accumulator.Len()
	i := 0
	for ; i+w <= len(x); i += w {
		accumulator = accumulator.Add(simd.LoadFloat32s(x[i:]))
	}
	// Current SDK widths are at most 16 float32 lanes. A larger future width
	// remains correct, with its scratch allocation visible in measurements.
	var scratch [16]float32
	lanes := scratch[:]
	if w > len(lanes) {
		lanes = make([]float32, w)
	}
	accumulator.Store(lanes)
	var total float32
	for _, v := range lanes[:w] {
		total = float32(total + v)
	}
	for ; i < len(x); i++ {
		total = float32(total + x[i])
	}
	return total
}
