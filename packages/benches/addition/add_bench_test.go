package addition_test

import (
	"fmt"
	"os"
	"testing"

	"simd"

	"github.com/akolybelnikov/go-simd/packages/benches"
	"github.com/akolybelnikov/go-simd/packages/benches/addition"
)

type addBenchmarkVariant struct {
	name string
	add  func([]float32, []float32, []float32) error
}

func BenchmarkAddition(b *testing.B) {
	var vector simd.Float32s
	sizes := benches.BenchmarkSizes(vector.Len())
	results := make(benches.BenchmarkResults, len(sizes))
	variants := []addBenchmarkVariant{
		{name: "Scalar", add: addition.AddScalar},
		{name: "SIMDPartial", add: addition.AddSIMD},
		{name: "SIMDTail", add: addition.AddSIMDVectorTail},
		{name: "SIMDParallel", add: addition.AddSIMDParallel},
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a, inputB, out := benches.MakeBenchmarkSlices(size)
			results[size] = make(map[string]float64, len(variants))

			for _, variant := range variants {
				variant := variant
				b.Run(variant.name, func(b *testing.B) {
					benches.Measure(b, size, out, results[size], variant.name, func() {
						for i := 0; i < b.N; i++ {
							if err := variant.add(a, inputB, out); err != nil {
								b.Fatal(err)
							}
						}
					})
				})
			}
		})
	}

	names := make([]string, len(variants))
	for i, variant := range variants {
		names[i] = variant.name
	}
	if err := benches.PrintComparison(os.Stdout, "Addition", sizes, results, names); err != nil {
		b.Fatal(err)
	}
}
