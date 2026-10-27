// A module-local example of all five F004 broadcast/predicate patterns.
package main

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func main() {
	x, y := []float32{-2, -1, 0, 1, 2}, []float32{10, 20, 30, 40, 50}
	out := make([]float32, len(x))
	for _, example := range []struct {
		name string
		run  func() error
	}{
		{"scale", func() error { return patterns.ScaleSIMD(x, out, 2) }},
		{"affine", func() error { return patterns.AffineSIMD(x, out, 2, 1) }},
		{"threshold", func() error { return patterns.ThresholdZeroSIMD(x, out, 0) }},
		{"selection", func() error { return patterns.SelectGTSIMD(x, y, out, 0) }},
		{"rangefilter", func() error { return patterns.RangeZeroSIMD(x, out, -2, 2) }},
	} {
		if err := example.run(); err != nil {
			panic(err)
		}
		fmt.Printf("%s: %v\n", example.name, out)
	}
}
