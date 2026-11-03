// Run inside this module with GOEXPERIMENT=simd.
package main

import (
	"fmt"
	"math"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func main() {
	var vector simd.Float32s
	w := vector.Len()
	fmt.Printf("vector gate: %s; lanes: %d\n", fmacap.Path(), w)
	x, y := make([]float32, 2*w), make([]float32, 2*w)
	x[0], y[0], x[w], y[w] = -1, 1, 1+0x1p-23, 1-0x1p-23
	unfused, err := reduction.DotScalar(x, y)
	must(err)
	fmt.Printf("unfused scalar bits: %08x\n", math.Float32bits(unfused))
	for _, k := range []struct {
		name string
		run  func([]float32, []float32) (float32, error)
	}{
		{"Scalar", reduction.DotFMAScalar}, {"SIMDStriped", reduction.DotFMASIMD}, {"SIMDStriped2", reduction.DotFMASIMD2}, {"SIMDStriped4", reduction.DotFMASIMD4},
	} {
		value, err := k.run(x, y)
		must(err)
		fmt.Printf("fused %s bits: %08x\n", k.name, math.Float32bits(value))
	}
	x, y = make([]float32, w), make([]float32, w)
	for i := range x {
		x[i], y[i] = math.SmallestNonzeroFloat32, -0.25
	}
	a, err := reduction.DotFMAScalar(x, y)
	must(err)
	b, err := reduction.DotFMASIMD(x, y)
	must(err)
	fmt.Printf("signed underflow: Scalar=%08x SIMDStriped=%08x\n", math.Float32bits(a), math.Float32bits(b))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
