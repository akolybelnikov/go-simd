// Run from this module: GOEXPERIMENT=simd go run ./examples/dot.
package main

import (
	"errors"
	"fmt"
	"math"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func main() {
	x, y := []float32{1, 2, 3}, []float32{4, 5, 6}
	a, err := reduction.DotScalar(x, y)
	must(err)
	b, err := reduction.DotSIMD(x, y)
	must(err)
	fmt.Printf("simple: Scalar=%g SIMDStriped=%g\n", a, b)
	var vector simd.Float32s
	w := vector.Len()
	x, y = make([]float32, 2*w), make([]float32, 2*w)
	for i := range y {
		y[i] = 1
	}
	x[0], x[1], x[w] = 16777216, 1, -16777216
	a, err = reduction.DotScalar(x, y)
	must(err)
	b, err = reduction.DotSIMD(x, y)
	must(err)
	fmt.Printf("lanes=%d; cancellation A: Scalar=%g SIMDStriped=%g\n", w, a, b)
	x[0], x[1], x[w] = 16777216, -16777216, 1
	a, err = reduction.DotScalar(x, y)
	must(err)
	b, err = reduction.DotSIMD(x, y)
	must(err)
	fmt.Printf("cancellation B: Scalar=%g SIMDStriped=%g\n", a, b)
	// Even a two-element dot exposes the separate product rounding contract.
	a, err = reduction.DotScalar([]float32{-1, 1 + 0x1p-23}, []float32{1, 1 - 0x1p-23})
	must(err)
	fmt.Printf("unfused result bits: 0x%08x\n", math.Float32bits(a))
	_, err = reduction.DotSIMD([]float32{1}, nil)
	fmt.Printf("length mismatch: %t\n", errors.Is(err, reduction.ErrLengthMismatch))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
