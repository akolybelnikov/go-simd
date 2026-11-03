package reduction

import (
	"errors"
	"simd"
)

// ErrLengthMismatch is returned before reading either input when lengths differ.
var ErrLengthMismatch = errors.New("reduction: input lengths differ")

// DotScalar rounds each product to float32, then adds left to right from +0.
// It does not fuse multiplication and addition or modify either input.
func DotScalar(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	var total float32
	for i, v := range x {
		product := float32(v * y[i])
		total = float32(total + product)
	}
	return total, nil
}

// DotSIMD rounds vector products before adding to one striped accumulator.
// It folds lanes low to high from +0, then adds rounded scalar-tail products.
// Length validation, extraction, final fold and tail belong to the kernel.
func DotSIMD(x, y []float32) (float32, error) {
	if len(x) != len(y) {
		return 0, ErrLengthMismatch
	}
	if len(x) == 0 {
		return 0, nil
	}
	var accumulator simd.Float32s
	w := accumulator.Len()
	i := 0
	for ; i+w <= len(x); i += w {
		product := simd.LoadFloat32s(x[i:]).Mul(simd.LoadFloat32s(y[i:]))
		accumulator = accumulator.Add(product)
	}
	// Current SDK widths fit on the stack; a larger future width remains
	// correct, with dynamic scratch visible in allocation counters.
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
		product := float32(x[i] * y[i])
		total = float32(total + product)
	}
	return total, nil
}
