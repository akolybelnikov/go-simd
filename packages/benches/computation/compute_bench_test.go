package computation_test

import (
	"fmt"
	"os"
	"testing"

	"simd"

	"github.com/akolybelnikov/go-simd/packages/benches"
	"github.com/akolybelnikov/go-simd/packages/benches/computation"
)

func BenchmarkComputation(b *testing.B) {
	var vector simd.Float32s
	sizes := benches.BenchmarkSizes(vector.Len())
	results := make(benches.BenchmarkResults, len(sizes))

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a, inputB, out := benches.MakeBenchmarkSlices(size)
			results[size] = make(map[string]float64, 2)

			b.Run("Scalar", func(b *testing.B) {
				benches.Measure(b, size, out, results[size], "Scalar", func() {
					for i := 0; i < b.N; i++ {
						if err := computation.ComputeHeavyScalar(a, inputB, out); err != nil {
							b.Fatal(err)
						}
					}
				})
			})
			b.Run("SIMD", func(b *testing.B) {
				benches.Measure(b, size, out, results[size], "SIMD", func() {
					for i := 0; i < b.N; i++ {
						if err := computation.ComputeHeavySIMD(a, inputB, out); err != nil {
							b.Fatal(err)
						}
					}
				})
			})
		})
	}

	if err := benches.PrintComparison(os.Stdout, "Computation", sizes, results, []string{"Scalar", "SIMD"}); err != nil {
		b.Fatal(err)
	}
}
