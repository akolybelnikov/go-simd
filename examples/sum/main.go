// Sum results can differ by fold order; timings and arithmetic outputs are
// separate evidence. Run this module-local example with GOEXPERIMENT=simd.
package main

import (
	"fmt"
	"math"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func main() {
	x := []float32{1, 2, 3}
	fmt.Printf("simple: Scalar=%g SIMDStriped=%g\n", reduction.SumScalar(x), reduction.SumSIMD(x))
	var vector simd.Float32s
	w := vector.Len()
	x = make([]float32, 2*w)
	x[0], x[1], x[w] = 16777216, 1, -16777216
	fmt.Printf("lanes=%d; cancellation A: Scalar=%g SIMDStriped=%g\n", w, reduction.SumScalar(x), reduction.SumSIMD(x))
	x[0], x[1], x[w] = 16777216, -16777216, 1
	fmt.Printf("cancellation B: Scalar=%g SIMDStriped=%g\n", reduction.SumScalar(x), reduction.SumSIMD(x))
	fmt.Printf("empty result bits: 0x%08x\n", math.Float32bits(reduction.SumSIMD(nil)))
}
