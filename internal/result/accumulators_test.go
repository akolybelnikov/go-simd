package result

import (
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestAccumulatorFoldIdentityAndNativeComparison(t *testing.T) {
	for _, test := range []struct {
		operation, benchmark string
		orders               []string
	}{
		{catalog.Sum, "BenchmarkSum", []string{"left-to-right-f32/v1", "striped-lane-fold-f32/v1", "striped-2-accumulators-lane-fold-f32/v1", "striped-4-accumulators-lane-fold-f32/v1"}},
		{catalog.Dot, "BenchmarkDot", []string{"left-to-right-unfused-dot-f32/v1", "striped-lane-unfused-dot-f32/v1", "striped-2-accumulators-unfused-dot-f32/v1", "striped-4-accumulators-unfused-dot-f32/v1"}},
	} {
		labels := []string{"Scalar", "SIMDStriped", "SIMDStriped2", "SIMDStriped4"}
		ids := []string{catalog.Scalar, catalog.SIMDStriped, catalog.SIMDStriped2, catalog.SIMDStriped4}
		counts := []int{0, 1, 2, 4}
		var raw strings.Builder
		for i, id := range ids {
			fold, err := ReductionFold(test.operation, id, 8)
			if err != nil || fold.VectorAccumulators != counts[i] || fold.OrderID != test.orders[i] || fold.NativeLanes.Value != 8 || !fold.NativeLanes.Known {
				t.Fatal("fold identity drift", fold, err)
			}
			raw.WriteString(test.benchmark + "/n=64/" + labels[i] + "-4 3 10 ns/op 0 result-bits 8 simd-lanes\n")
		}
		run, err := ReadLegacy(strings.NewReader(raw.String()), LegacyOptions{RunID: "accumulators"})
		if err != nil {
			t.Fatal(err)
		}
		for i, o := range run.Observations {
			if o.Key.Case.Reduction.VectorAccumulators != counts[i] || o.Key.Case.Reduction.OrderID != test.orders[i] {
				t.Fatal("native accumulator identity lost")
			}
		}
		summaries, err := Summarize(run.Observations)
		if err != nil {
			t.Fatal(err)
		}
		table, err := Compare("accumulators", summaries)
		if err != nil || len(table.Columns) != 4 {
			t.Fatal("strategy comparison failed", err)
		}
		for i, column := range table.Columns {
			if column.Label != labels[i] {
				t.Fatal("column ordering drift")
			}
		}
		summaries[3].Key.Case.Reduction.NativeLanes.Value = 4
		if _, err := Compare("accumulators", summaries); err == nil {
			t.Fatal("mixed widths pooled")
		}
	}
}
