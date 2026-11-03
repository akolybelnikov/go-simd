// Run from this module with GOEXPERIMENT=simd.
package main

import (
	"fmt"
	"math"

	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func main() {
	fmt.Println("FMA32 SIMD path:", patterns.FMAPath())
	alpha, x, beta := float32(1+0x1p-23), []float32{1 - 0x1p-23}, float32(-1)
	unfused, fusedScalar, fusedSIMD := make([]float32, 1), make([]float32, 1), make([]float32, 1)
	must(patterns.AffineScalar(x, unfused, alpha, beta))
	must(patterns.AffineFMAScalar(x, fusedScalar, alpha, beta))
	must(patterns.AffineFMASIMD(x, fusedSIMD, alpha, beta))
	fmt.Printf("unfused bits: %08x; fused scalar: %08x; fused SIMD: %08x\n", math.Float32bits(unfused[0]), math.Float32bits(fusedScalar[0]), math.Float32bits(fusedSIMD[0]))
	beta = 16777218
	naive := float32(math.FMA(float64(alpha), float64(x[0]), float64(beta)))
	fmt.Printf("double-rounding example: widened=%g corrected=%g\n", naive, fma32.FMA(alpha, x[0], beta))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
