package benches

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"
	"text/tabwriter"
)

// BenchmarkResults stores nanoseconds per operation by input size and variant.
type BenchmarkResults map[int]map[string]float64

var benchmarkSink float32

// BadInputError is shared by operation packages when slice lengths disagree.
var BadInputError = errors.New("input slices must have the same length and match the output slice length")

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
	a = make([]float32, size)
	inputB = make([]float32, size)
	out = make([]float32, size)
	for i := range a {
		a[i] = float32(i%100) / 10
		inputB[i] = float32(i%50-25) / 10
	}
	return a, inputB, out
}

// Measure runs a benchmark closure and records its average duration.
func Measure(
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

// PrintComparison writes a formatted time comparison table.
func PrintComparison(
	w io.Writer,
	title string,
	sizes []int,
	results BenchmarkResults,
	names []string,
) error {
	if _, err := fmt.Fprintf(w, "\n%s benchmark comparison (ns/op; change vs scalar)\n", title); err != nil {
		return err
	}
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprint(table, "ELEMENTS\t"); err != nil {
		return err
	}
	for _, name := range names {
		if _, err := fmt.Fprintf(table, "%s ns/op\t", name); err != nil {
			return err
		}
		if name != "Scalar" {
			if _, err := fmt.Fprintf(table, "%s change\t", name); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(table); err != nil {
		return err
	}

	for _, size := range sizes {
		timings, ok := results[size]
		if !ok {
			return fmt.Errorf("missing benchmark results for size %d", size)
		}
		scalar, ok := timings["Scalar"]
		if !ok {
			return fmt.Errorf("missing scalar timing for size %d", size)
		}
		if _, err := fmt.Fprintf(table, "%d\t", size); err != nil {
			return err
		}
		for _, name := range names {
			elapsed, ok := timings[name]
			if !ok {
				return fmt.Errorf("missing %s timing for size %d", name, size)
			}
			if _, err := fmt.Fprintf(table, "%.2f\t", elapsed); err != nil {
				return err
			}
			if name != "Scalar" {
				if _, err := fmt.Fprintf(table, "%+.1f%%\t", (elapsed-scalar)/scalar*100); err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintln(table); err != nil {
			return err
		}
	}
	return table.Flush()
}
