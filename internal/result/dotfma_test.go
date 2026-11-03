package result

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
)

func TestFusedDotCasePathsAndNativeMetrics(t *testing.T) {
	for _, strategy := range []string{catalog.Scalar, catalog.SIMDStriped, catalog.SIMDStriped2, catalog.SIMDStriped4} {
		for _, vectorPath := range []string{fma32.SoftwarePath, fma32.VectorPath} {
			for _, n := range []int{1, 7, 8, 9, 16, 17} {
				i, err := NewDotFMAIdentity(strategy, vectorPath, 8, n)
				if err != nil {
					t.Fatal(err)
				}
				want := fma32.SoftwarePath
				if strategy != catalog.Scalar {
					if vectorPath == fma32.SoftwarePath {
						want = SoftwareStripedFMAPath
					} else if n < 8 {
						want = SoftwareTailFMAPath
					} else if n%8 != 0 {
						want = VectorTailFMAPath
					} else {
						want = fma32.VectorPath
					}
				}
				if i.PathID != want || i.VectorPathID != vectorPath || i.RoundingID != FusedDotRoundingID || i.NativeLanes.Value != 8 {
					t.Fatal("wrong case path", i)
				}
				code, err := DotFMAPathCode(i.PathID)
				if err != nil {
					t.Fatal(err)
				}
				label := "Scalar"
				switch strategy {
				case catalog.SIMDStriped:
					label = "SIMDStriped"
				case catalog.SIMDStriped2:
					label = "SIMDStriped2"
				case catalog.SIMDStriped4:
					label = "SIMDStriped4"
				}
				flag := 0
				if vectorPath == fma32.VectorPath {
					flag = 1
				}
				raw := fmt.Sprintf("BenchmarkDotFMA/n=%d/%s-4 3 10 ns/op %d fma-path %d fma-vector 8 simd-lanes 2826960896 result-bits", n, label, code, flag)
				run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "fused-dot"})
				if err != nil || run.HarnessID != LegacyDotFMAHarnessID || run.Observations[0].Key.Case.FMA != i || *run.Observations[0].ScalarResultBits != 2826960896 {
					t.Fatal("fused-dot native facts lost", err)
				}
			}
		}
	}
	for _, metrics := range []string{"8 simd-lanes 0 result-bits", "8 simd-lanes 0 result-bits 1 fma-vector 1 fma-path", "8 simd-lanes 0 result-bits 0 fma-vector 2.5 fma-path", "8 simd-lanes 0 result-bits 2 fma-vector 2 fma-path", "8 simd-lanes 0 result-bits 0 fma-vector 5 fma-path"} {
		if _, err := ReadLegacy(strings.NewReader("BenchmarkDotFMA/n=1/SIMDStriped-4 1 10 ns/op "+metrics), LegacyOptions{RunID: "bad"}); err == nil {
			t.Fatal("bad fused-dot metrics accepted", metrics)
		}
	}
	raw := "BenchmarkDotFMA/n=9/Scalar-4 1 10 ns/op 0 fma-path 1 fma-vector 8 simd-lanes 0 result-bits\nBenchmarkDotFMA/n=9/SIMDStriped-4 1 5 ns/op 3 fma-path 1 fma-vector 8 simd-lanes 0 result-bits"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "compare"})
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := Summarize(run.Observations)
	if err != nil {
		t.Fatal(err)
	}
	if table, err := Compare("fused", summaries); err != nil || table.BaselinePolicy != "specified-fused-float32-dot-folds/v1" {
		t.Fatal("fused-dot comparison", err)
	}
	summaries[1].Key.Case.FMA.VectorPathID = fma32.SoftwarePath
	if _, err := Compare("mixed", summaries); err == nil {
		t.Fatal("mixed process gates pooled")
	}
	for _, test := range []struct {
		id, want string
		count    int
	}{
		{catalog.Scalar, "left-to-right-fma-dot-f32/v1", 0}, {catalog.SIMDStriped, "striped-lane-fma-dot-f32/v1", 1}, {catalog.SIMDStriped2, "striped-2-accumulators-fma-dot-f32/v1", 2}, {catalog.SIMDStriped4, "striped-4-accumulators-fma-dot-f32/v1", 4},
	} {
		f, err := ReductionFold(catalog.DotFMA, test.id, 8)
		if err != nil || f.OrderID != test.want || f.VectorAccumulators != test.count {
			t.Fatal("fused fold identity drift", f)
		}
	}
}
