package result

import (
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/fma32"
)

func TestFMAChildIdentityAndStrictMetrics(t *testing.T) {
	raw := "BenchmarkAffineFMA/n=8/Scalar-4 2 10 ns/op 0 fma-path 1 fma-vector 8 simd-lanes 0 last-result-bits\nBenchmarkAffineFMA/n=8/SIMDPartial-4 2 5 ns/op 1 fma-path 1 fma-vector 8 simd-lanes 0 last-result-bits"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "fma"})
	if err != nil || run.HarnessID != LegacyFMAHarnessID {
		t.Fatal("FMA native import", err)
	}
	if run.Observations[0].Key.Case.FMA.PathID != fma32.SoftwarePath || run.Observations[1].Key.Case.FMA.PathID != fma32.VectorPath || run.Observations[0].ArrayResultLastBits == nil {
		t.Fatal("FMA strategy/path facts lost")
	}
	summaries, err := Summarize(run.Observations)
	if err != nil {
		t.Fatal(err)
	}
	if table, err := Compare("fma", summaries); err != nil || table.BaselinePolicy != "fused-software-reference/v1" {
		t.Fatal("FMA comparison", err)
	}
	summaries[1].Key.Case.FMA.VectorPathID = fma32.SoftwarePath
	if _, err := Compare("fma", summaries); err == nil {
		t.Fatal("mixed FMA process paths pooled")
	}
	for _, metrics := range []string{"", "0 fma-path 8 simd-lanes 0 last-result-bits", "0 fma-path 0 fma-vector 8.5 simd-lanes 0 last-result-bits", "1 fma-path 1 fma-vector 8 simd-lanes 0 last-result-bits", "0 fma-path 2 fma-vector 8 simd-lanes 0 last-result-bits", "0 fma-path 0 fma-vector 8 simd-lanes 4294967296 last-result-bits"} {
		if _, err := ReadLegacy(strings.NewReader("BenchmarkAffineFMA/n=8/Scalar-4 1 10 ns/op "+metrics), LegacyOptions{RunID: "bad"}); err == nil {
			t.Fatal("invalid FMA metrics accepted", metrics)
		}
	}
}
