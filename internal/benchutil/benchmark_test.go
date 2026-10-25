package benchutil

import (
	"bytes"
	"reflect"
	"simd"
	"slices"
	"strings"
	"testing"
)

func TestBenchmarkSizesIncludesVectorBoundaries(t *testing.T) {
	var vector simd.Float32s
	width := vector.Len()
	got := BenchmarkSizes(width)

	for _, size := range []int{1, 1 << 24, width - 1, width, width + 1, 2*width - 1, 2 * width, 2*width + 1} {
		if !slices.Contains(got, size) {
			t.Errorf("BenchmarkSizes() missing %d", size)
		}
	}
	if !reflect.DeepEqual(got, slices.Compact(slicesSorted(got))) {
		t.Fatal("BenchmarkSizes() should return sorted unique sizes")
	}
}

func TestMakeBenchmarkSlices(t *testing.T) {
	a, b, out := MakeBenchmarkSlices(4)
	if len(a) != 4 || len(b) != 4 || len(out) != 4 {
		t.Fatalf("MakeBenchmarkSlices() lengths = %d, %d, %d; want 4", len(a), len(b), len(out))
	}
	if a[0] != 0 || b[0] != -2.5 || a[3] != 0.3 || b[3] != -2.2 {
		t.Errorf("unexpected sample input values: a=%v b=%v", a, b)
	}
}

func TestPrintComparison(t *testing.T) {
	var output bytes.Buffer
	err := PrintComparison(&output, "Addition", []int{1}, BenchmarkResults{
		1: {"Scalar": 10, "SIMD": 8},
	}, []string{"Scalar", "SIMD"})
	if err != nil {
		t.Fatalf("PrintComparison() error = %v", err)
	}
	for _, want := range []string{
		"Addition benchmark comparison",
		"ELEMENTS",
		"1",
		"10.00",
		"8.00",
		"-20.0%",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("PrintComparison() output missing %q: %s", want, output.String())
		}
	}
}

func TestPrintComparisonAllowsMissingTiming(t *testing.T) {
	var out bytes.Buffer
	err := PrintComparison(&out, "Addition", []int{4}, BenchmarkResults{4: {"SIMD": 8}}, []string{"Scalar", "SIMD"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Inf") || !strings.Contains(out.String(), "—") {
		t.Fatalf("unexpected table: %s", out.String())
	}
}

func slicesSorted(values []int) []int {
	result := append([]int(nil), values...)
	slices.Sort(result)
	return result
}
