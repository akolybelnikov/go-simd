package result

import (
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestClampNativeFactsAndContexts(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/n=9/SIMDPartial", "BenchmarkClamp/n=9/SIMDMinMax", 1)
	options := input.DefaultOptions()
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "clamp", GeneratedInputs: true, InputOptions: &options})
	if err != nil || run.HarnessID != LegacyClampHarnessID || len(run.Observations) != 1 {
		t.Fatal(run, err)
	}
	o := run.Observations[0]
	op, _ := catalog.ByGroup("clamp")
	if o.Key.Case.StrategyID != catalog.SIMDMinMax || o.Key.Case.Clamp.PolicyID != ClampPolicyID || o.Key.Case.Clamp.NativeLanes.Value != 8 || o.ArrayResultSHA256 != "00010203ffffffff000000001234567889abcdef800000007fffffff00112233" || o.ByteResultSHA256 != "" || o.Key.Case.Input.GeneratorID != input.DefaultOptions().Generator(op) || o.Key.Case.Input.SHA256 != input.FingerprintForOperation(op, 9, input.DefaultOptions()) {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, " 8 simd-lanes", "", 1), strings.Replace(raw, " 8 simd-lanes", " 1.5 simd-lanes", 1), strings.Replace(raw, "0 output-sha256-2", "", 1), strings.Replace(raw, "0 output-sha256-2", "0.5 output-sha256-2", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt clamp facts accepted")
		}
	}
	history, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "history"})
	if e != nil || history.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("historical copy fingerprint invented", e)
	}
	if _, e := NewClampIdentity(0); e == nil {
		t.Fatal("zero width")
	}
	for _, field := range []string{"width", "bits"} {
		a, b := o, o
		b.Key.Case.StrategyID = catalog.Scalar
		b.Label = "Scalar"
		if field == "width" {
			b.Key.Case.Clamp.NativeLanes.Value = 4
		} else {
			b.Key.Case.Clamp.PolicyID = "numeric-conversion/v1"
		}
		s, e := Summarize([]Observation{a, b})
		if e != nil {
			t.Fatal(e)
		}
		if _, e := Compare("mixed", s); e == nil {
			t.Fatal("clamp contexts pooled", field)
		}
	}
}
