package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestComparison(t *testing.T) {
	raw := `goos: windows
BenchmarkAddition/n=1/Scalar-16 100 10 ns/op 0 B/op 0 allocs/op
BenchmarkAddition/n=1/Scalar-16 100 30 ns/op 0 B/op 0 allocs/op
BenchmarkAddition/n=1/Scalar-16 100 20 ns/op 0 B/op 0 allocs/op
BenchmarkAddition/n=1/SIMDPartial-16 100 10 ns/op 0 B/op 0 allocs/op
`
	out, err := comparison(raw, "Addition")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"20.00", "10.00", "-50.0%"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %s in %s", want, out)
		}
	}
}
func TestComparisonFiltered(t *testing.T) {
	out, err := comparison("BenchmarkAddition/n=1/SIMDPartial-16 1 10 ns/op 0 B/op 0 allocs/op", "Addition")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "10.00") || strings.Contains(out, "Inf") {
		t.Fatal(out)
	}
}
func TestComparisonEmpty(t *testing.T) {
	if _, err := comparison("PASS", "Addition"); err == nil {
		t.Fatal("expected missing measurements error")
	}
}

// Keep the original report shape: one row per size and a percentage column
// beside each SIMD timing, regardless of measurement order or repetitions.
func TestComparisonMultiSizeTable(t *testing.T) {
	raw := `BenchmarkAddition/n=16/SIMDTail-16 100 5 ns/op
BenchmarkAddition/n=8/Scalar-16 100 10 ns/op
BenchmarkAddition/n=16/Scalar-16 100 20 ns/op
BenchmarkAddition/n=8/SIMDPartial-16 100 8 ns/op
BenchmarkAddition/n=16/SIMDPartial-16 100 10 ns/op
BenchmarkAddition/n=8/SIMDTail-16 100 12 ns/op
BenchmarkAddition/n=16/SIMDParallel-16 100 30 ns/op
BenchmarkAddition/n=8/SIMDParallel-16 100 40 ns/op
`
	out, err := comparison(raw, "Addition")
	if err != nil {
		t.Fatal(err)
	}
	var rows [][]string
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && (fields[0] == "8" || fields[0] == "16") {
			rows = append(rows, fields)
		}
	}
	want := [][]string{
		{"8", "10.00", "8.00", "-20.0%", "12.00", "+20.0%", "40.00", "+300.0%"},
		{"16", "20.00", "10.00", "-50.0%", "5.00", "-75.0%", "30.00", "+50.0%"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows = %v, want %v\n%s", rows, want, out)
	}
	for _, column := range []string{"Scalar ns/op", "SIMDPartial ns/op", "SIMDPartial change", "SIMDTail ns/op", "SIMDTail change", "SIMDParallel ns/op", "SIMDParallel change"} {
		if !strings.Contains(out, column) {
			t.Errorf("missing column %q", column)
		}
	}
}
