package result

import (
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestDotNativeMetricsAndComparisonIdentity(t *testing.T) {
	raw := "BenchmarkDot/n=64/Scalar-4 3 10 ns/op 0 result-bits 8 simd-lanes\nBenchmarkDot/n=64/SIMDStriped-4 3 5 ns/op 1065353216 result-bits 8 simd-lanes"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "dot"})
	if err != nil || run.HarnessID != LegacyDotHarnessID {
		t.Fatal("dot native import failed", err)
	}
	if *run.Observations[1].ScalarResultBits != 1065353216 || run.Observations[1].Key.Case.Reduction.OrderID != "striped-lane-unfused-dot-f32/v1" {
		t.Fatal("dot facts lost")
	}
	summaries, err := Summarize(run.Observations)
	if err != nil {
		t.Fatal(err)
	}
	table, err := Compare("dot", summaries)
	if err != nil || table.BaselinePolicy != "specified-unfused-float32-dot-folds/v1" {
		t.Fatal("dot baseline policy lost", err)
	}
	summaries[1].Key.Case.Reduction.NativeLanes.Value = 4
	if _, err := Compare("dot", summaries); err == nil {
		t.Fatal("mixed dot widths pooled")
	}
	for _, line := range []string{
		"BenchmarkDot/n=8/Scalar-4 1 10 ns/op",
		"BenchmarkDot/n=8/SIMDStriped-4 1 10 ns/op 8 simd-lanes",
		"BenchmarkDot/n=8/Scalar-4 1 10 ns/op 0.5 result-bits 8 simd-lanes",
		"BenchmarkDot/n=8/Scalar-4 1 10 ns/op 4294967296 result-bits 8 simd-lanes",
		"BenchmarkDot/n=8/Scalar-4 1 10 ns/op 0 result-bits 0 simd-lanes",
		"BenchmarkDot/n=8/Scalar-4 1 10 ns/op 0 result-bits 2.5 simd-lanes",
	} {
		if _, err := ReadLegacy(strings.NewReader(line), LegacyOptions{RunID: "bad"}); err == nil {
			t.Fatal("bad dot metrics accepted", line)
		}
	}
	if _, err := ReductionFold(catalog.Dot, "unknown", 8); err == nil {
		t.Fatal("unknown dot strategy accepted")
	}
	if _, err := ReductionFold("unknown", catalog.Scalar, 8); err == nil {
		t.Fatal("unknown reduction accepted")
	}
}
