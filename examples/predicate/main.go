// Run inside this module with GOEXPERIMENT=simd.
package main

import (
	"fmt"
	"math"

	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func main() {
	x := []float32{-2, math.Float32frombits(0x80000000), 0, 1, 2, float32(math.NaN()), float32(math.Inf(1))}
	scalar, vector := make([]uint8, len(x)), make([]uint8, len(x))
	if err := patterns.PredicateGTScalar(x, scalar, 0); err != nil {
		panic(err)
	}
	if err := patterns.PredicateGTSIMD(x, vector, 0); err != nil {
		panic(err)
	}
	fmt.Println("x > 0, Scalar:    ", scalar)
	fmt.Println("x > 0, SIMDPartial:", vector)
	if err := patterns.PredicateGTSIMD([]float32{-2}, vector[:1], -1); err != nil {
		panic(err)
	}
	fmt.Println("one valid false lane at threshold -1:", vector[:1])
}
