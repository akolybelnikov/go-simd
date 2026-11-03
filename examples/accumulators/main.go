// Run from the module: GOEXPERIMENT=simd go run ./examples/accumulators.
package main

import (
	"fmt"
	"simd"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func main() {
	var vector simd.Float32s
	w := vector.Len()
	fmt.Printf("float32 lanes: %d\n", w)
	for _, layout := range []int{1, 2} {
		x, y := make([]float32, 8*w), make([]float32, 8*w)
		for i := range y {
			y[i] = 1
		}
		x[0], x[layout*w], x[2*layout*w] = 16777216, 1, -16777216
		fmt.Printf("layout %d sum: Scalar=%g SIMDStriped=%g SIMDStriped2=%g SIMDStriped4=%g\n", layout, reduction.SumScalar(x), reduction.SumSIMD(x), reduction.SumSIMD2(x), reduction.SumSIMD4(x))
		one, err := reduction.DotSIMD(x, y)
		must(err)
		two, err := reduction.DotSIMD2(x, y)
		must(err)
		four, err := reduction.DotSIMD4(x, y)
		must(err)
		fmt.Printf("layout %d dot: SIMDStriped=%g SIMDStriped2=%g SIMDStriped4=%g\n", layout, one, two, four)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
