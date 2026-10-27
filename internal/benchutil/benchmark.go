package benchutil

import (
	"slices"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
)

var benchmarkSink float32

// BenchmarkSizes returns powers of two and sizes around two SIMD vector widths.
func BenchmarkSizes(width int) []int {
	sizes := make([]int, 0, 32)
	for size := 1; size <= 1<<24; size *= 2 {
		sizes = append(sizes, size)
	}
	for _, size := range []int{
		width - 1,
		width,
		width + 1,
		2*width - 1,
		2 * width,
		2*width + 1,
	} {
		if size > 0 {
			sizes = append(sizes, size)
		}
	}
	slices.Sort(sizes)
	return slices.Compact(sizes)
}

// MakeBenchmarkSlices creates deterministic inputs and a reusable output slice.
func MakeBenchmarkSlices(size int) (a, inputB, out []float32) {
	return input.MakeFloat32Slices(size)
}

// Measure runs a benchmark with preallocated inputs and includes call overhead.
func Measure(b *testing.B, size int, out []float32, run func() error) {
	b.ReportAllocs()
	b.SetBytes(int64(size) * 12)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := run(); err != nil {
			b.Fatal(err)
		}
	}
	if size > 0 {
		benchmarkSink = out[size-1]
	}
}
