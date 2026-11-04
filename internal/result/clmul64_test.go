package result

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestClmul64NativeFacts(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkClmul64/", 1)
	raw = strings.TrimSpace(raw) + " 1 clmul-hardware\n"
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "clmul", GeneratedInputs: true})
	if e != nil || run.HarnessID != LegacyClmul64HarnessID {
		t.Fatal(run, e)
	}
	o := run.Observations[0]
	op, _ := catalog.ByGroup("clmul64")
	if o.Key.Case.Input.SHA256 != input.FingerprintForOperation(op, 9, input.DefaultOptions()) || len(o.ArrayResultSHA256) != 64 || o.Key.Case.Vector.NativeLanes.Value != 8 || o.Key.Case.Vector.CarrylessCapabilityID != CarrylessCapability(true) {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, "1 clmul-hardware", "", 1), strings.Replace(raw, "1 clmul-hardware", "0.5 clmul-hardware", 1), strings.Replace(raw, "8 simd-lanes", "", 1), strings.Replace(raw, "8 simd-lanes", "1.5 simd-lanes", 1), strings.Replace(raw, "0 output-sha256-2", "", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt carryless metrics")
		}
	}
	old, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "old"})
	if e != nil || old.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("historical input invented")
	}
}

func TestCarrylessCapabilitySeparatesComparisons(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkClmul64/", 1)
	raw = strings.TrimSpace(raw) + " 1 clmul-hardware\n"
	r, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "x"})
	if e != nil {
		t.Fatal(e)
	}
	a, b := r.Observations[0], r.Observations[0]
	b.Key.Case.StrategyID = catalog.Scalar
	b.Label = "Scalar"
	b.Key.Case.Vector.CarrylessCapabilityID = CarrylessCapability(false)
	ss, e := Summarize([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Compare("mixed", ss); e == nil {
		t.Fatal("capability mixed")
	}
}
