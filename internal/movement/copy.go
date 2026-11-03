// Package movement contains complete data-movement experiments.
package movement

import (
	"errors"
	"simd"
)

var ErrLengthMismatch = errors.New("movement: slice lengths differ")

// CopyScalar copies exact float32 bits using a source loop. The label describes
// source structure, not proof of scalar machine instructions. Storage must be
// disjoint or an exact alias; partial overlap is outside the shared contract.
func CopyScalar(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i, v := range x {
		out[i] = v
	}
	return nil
}

// CopyBuiltin wraps Go copy with the same strict length/alias contract as the
// experiment's other kernels. Compiler/runtime lowering determines execution.
func CopyBuiltin(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	copy(out, x)
	return nil
}

// CopySIMD uses partial-vector loads/stores, including the final short vector.
func CopySIMD(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		v.StorePart(out[i:])
		i += n
	}
	return nil
}

// CopySIMDVectorTail uses full vectors followed by scalar assignments.
func CopySIMDVectorTail(x, out []float32) error {
	if len(x) != len(out) {
		return ErrLengthMismatch
	}
	var vector simd.Float32s
	w := vector.Len()
	i := 0
	for ; i+w <= len(x); i += w {
		simd.LoadFloat32s(x[i:]).Store(out[i:])
	}
	for ; i < len(x); i++ {
		out[i] = x[i]
	}
	return nil
}
