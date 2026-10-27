package benchutil

import (
	"reflect"
	"simd"
	"slices"
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

func TestLegacyBenchmarkSweep(t *testing.T) {
	// Fixed width makes the complete historical sweep deterministic on every
	// development target, including targets with a different native SIMD width.
	want := []int{
		1, 2, 4, 7, 8, 9, 15, 16, 17, 32, 64, 128, 256, 512, 1024,
		2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144,
		524288, 1048576, 2097152, 4194304, 8388608, 16777216,
	}
	if got := BenchmarkSizes(8); !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy sweep = %v, want %v", got, want)
	}
}

func TestLegacyGeneratedFixtureWraparound(t *testing.T) {
	a, b, out := MakeBenchmarkSlices(101)
	for _, tt := range []struct {
		index int
		a, b  float32
	}{
		{0, 0, -2.5},
		{1, 0.1, -2.4},
		{49, 4.9, 2.4},
		{50, 5, -2.5},
		{99, 9.9, 2.4},
		{100, 0, -2.5},
	} {
		if a[tt.index] != tt.a || b[tt.index] != tt.b {
			t.Errorf("fixture[%d] = (%v, %v), want (%v, %v)", tt.index, a[tt.index], b[tt.index], tt.a, tt.b)
		}
	}
	for i, value := range out {
		if value != 0 {
			t.Fatalf("output[%d] is not initially zero: %v", i, value)
		}
	}
}

func slicesSorted(values []int) []int {
	result := append([]int(nil), values...)
	slices.Sort(result)
	return result
}
