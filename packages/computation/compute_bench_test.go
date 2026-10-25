package computation_test

import (
	"fmt"
	"os"
	"slices"
	"testing"
	"text/tabwriter"

	"simd"

	"github.com/akolybelnikov/go-simd/packages/computation"
)

var benchmarkSink float32

type benchmarkResults map[int]map[string]float64

func BenchmarkComputation(b *testing.B) {
	var vector simd.Float32s
	sizes := benchmarkSizes(vector.Len())
	results := make(benchmarkResults, len(sizes))

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a, inputB, out := makeBenchmarkSlices(size)
			results[size] = make(map[string]float64, 2)

			b.Run("Scalar", func(b *testing.B) {
				measure(b, size, out, results[size], "Scalar", func() {
					for i := 0; i < b.N; i++ {
						if err := computation.ComputeHeavyScalar(a, inputB, out); err != nil {
							b.Fatal(err)
						}
					}
				})
			})
			b.Run("SIMD", func(b *testing.B) {
				measure(b, size, out, results[size], "SIMD", func() {
					for i := 0; i < b.N; i++ {
						if err := computation.ComputeHeavySIMD(a, inputB, out); err != nil {
							b.Fatal(err)
						}
					}
				})
			})
		})
	}

	printComparison(b, "Computation", sizes, results, []string{"Scalar", "SIMD"})
}

func benchmarkSizes(width int) []int {
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

func makeBenchmarkSlices(size int) (a, inputB, out []float32) {
	a = make([]float32, size)
	inputB = make([]float32, size)
	out = make([]float32, size)
	for i := range a {
		a[i] = float32(i%100) / 10
		inputB[i] = float32(i%50-25) / 10
	}
	return a, inputB, out
}

func measure(
	b *testing.B,
	size int,
	out []float32,
	results map[string]float64,
	name string,
	run func(),
) {
	b.ReportAllocs()
	b.SetBytes(int64(size * 3 * 4))
	b.ResetTimer()
	run()
	if size > 0 {
		benchmarkSink = out[size-1]
	}
	b.StopTimer()
	results[name] = float64(b.Elapsed()) / float64(b.N)
}

func printComparison(
	b *testing.B,
	title string,
	sizes []int,
	results benchmarkResults,
	names []string,
) {
	fmt.Fprintf(os.Stdout, "\n%s benchmark comparison (ns/op; change vs scalar)\n", title)
	table := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprint(table, "ELEMENTS\t")
	for _, name := range names {
		fmt.Fprintf(table, "%s ns/op\t", name)
		if name != "Scalar" {
			fmt.Fprintf(table, "%s change\t", name)
		}
	}
	fmt.Fprintln(table)

	for _, size := range sizes {
		timings := results[size]
		scalar, ok := timings["Scalar"]
		if !ok {
			continue
		}
		fmt.Fprintf(table, "%d\t", size)
		for _, name := range names {
			elapsed, ok := timings[name]
			if !ok {
				b.Fatalf("missing %s timing for size %d", name, size)
			}
			fmt.Fprintf(table, "%.2f\t", elapsed)
			if name != "Scalar" {
				fmt.Fprintf(table, "%+.1f%%\t", (elapsed-scalar)/scalar*100)
			}
		}
		fmt.Fprintln(table)
	}
	if err := table.Flush(); err != nil {
		b.Fatal(err)
	}
}
