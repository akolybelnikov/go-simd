// Clamp demonstrates exact float32 output, including signed zeros/NaN payloads.
package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
)

func main() {
	x := []float32{-2, math.Float32frombits(0x80000000), 0, 0.5, 2, math.Float32frombits(0x7fc12345), math.Float32frombits(0x7f812345)}
	for _, bounds := range [][2]float32{{-1, 1}, {0, math.Float32frombits(0x80000000)}} {
		fmt.Printf("bounds: %08x %08x\n", math.Float32bits(bounds[0]), math.Float32bits(bounds[1]))
		for _, impl := range []struct {
			name string
			run  func([]float32, []float32, float32, float32) error
		}{{"Scalar", patterns.ClampScalar}, {"SIMDPartial", patterns.ClampSIMD}, {"SIMDMinMax", patterns.ClampSIMDMinMax}} {
			out := make([]float32, len(x))
			if err := impl.run(x, out, bounds[0], bounds[1]); err != nil {
				panic(err)
			}
			fmt.Printf("%-12s", impl.name)
			for _, v := range out {
				fmt.Printf(" %08x", math.Float32bits(v))
			}
			fmt.Println()
		}
	}
}
