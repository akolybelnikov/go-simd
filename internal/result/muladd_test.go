package result

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestMulAddStrictNativeFacts(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkMulAdd/", 1)
	raw = strings.TrimSpace(raw) + " 1 fma-vector 1 fma-path\n"
	options := input.DefaultOptions()
	options.FMAInput = "rounding"
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "muladd", GeneratedInputs: true, InputOptions: &options})
	if e != nil || run.HarnessID != LegacyMulAddHarnessID || len(run.Observations) != 1 {
		t.Fatal(run, e)
	}
	o := run.Observations[0]
	op, _ := catalog.ByGroup("muladd")
	if o.ArrayResultSHA256 != "00010203ffffffff000000001234567889abcdef800000007fffffff00112233" || o.Key.Case.FMA.PathID != fma32.VectorPath || o.Key.Case.Input.SHA256 != input.FingerprintForOperation(op, 9, options) {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, " 8 simd-lanes", "", 1), strings.Replace(raw, " 8 simd-lanes", " 1.5 simd-lanes", 1), strings.Replace(raw, " 1 fma-path", "", 1), strings.Replace(raw, " 1 fma-path", " 0 fma-path", 1), strings.Replace(raw, " 1 fma-vector", " 0 fma-vector", 1), strings.Replace(raw, "0 output-sha256-2", "", 1), strings.Replace(raw, "0 output-sha256-2", "0.5 output-sha256-2", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt muladd metrics accepted", bad)
		}
	}
	historical, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "historical", GeneratedInputs: true})
	if e != nil || historical.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("invented historical triple", e)
	}
}
