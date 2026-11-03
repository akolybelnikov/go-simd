package result

import (
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestReductionComparisonRequiresConsistentWidth(t *testing.T) {
	raw := "BenchmarkSum/n=8/Scalar-4 1 10 ns/op 0 result-bits 8 simd-lanes\nBenchmarkSum/n=8/SIMDStriped-4 1 5 ns/op 0 result-bits 8 simd-lanes"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "sum"})
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := Summarize(run.Observations)
	if err != nil {
		t.Fatal(err)
	}
	table, err := Compare("sum", summaries)
	if err != nil || table.BaselinePolicy != "specified-float32-folds/v1" {
		t.Fatalf("sum comparison: %+v %v", table, err)
	}
	summaries[1].Key.Case.Reduction.NativeLanes.Value = 4
	if _, err := Compare("sum", summaries); err == nil {
		t.Fatal("mixed native fold widths pooled")
	}
	for _, line := range []string{
		"BenchmarkSum/n=8/Scalar-4 1 10 ns/op",
		"BenchmarkSum/n=8/SIMDStriped-4 1 10 ns/op 8 simd-lanes",
		"BenchmarkSum/n=8/Scalar-4 1 10 ns/op 0 result-bits 2.5 simd-lanes",
		"BenchmarkSum/n=8/Scalar-4 1 10 ns/op 4294967296 result-bits 8 simd-lanes",
	} {
		if _, err := ReadLegacy(strings.NewReader(line), LegacyOptions{RunID: "bad"}); err == nil {
			t.Fatal("missing/malformed fold metrics accepted")
		}
	}
	if _, err := SumFold(catalog.SIMDStriped, 0); err == nil {
		t.Fatal("unknown width accepted")
	}
}
