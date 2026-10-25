package addition_test

import (
	"fmt"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/benchutil"
)

func BenchmarkAddition(b *testing.B) {
	var vector simd.Float32s
	variants := []struct {
		name string
		run  func([]float32, []float32, []float32) error
	}{
		{name: "Scalar", run: addition.AddScalar},
		{name: "SIMDPartial", run: addition.AddSIMD},
		{name: "SIMDTail", run: addition.AddSIMDVectorTail},
		{name: "SIMDParallel", run: addition.AddSIMDParallel},
	}
	for _, size := range benchutil.BenchmarkSizes(vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a, inputB, out := benchutil.MakeBenchmarkSlices(size)
			for _, variant := range variants {
				b.Run(variant.name, func(b *testing.B) {
					benchutil.Measure(b, size, out, func() error { return variant.run(a, inputB, out) })
				})
			}
		})
	}
}
