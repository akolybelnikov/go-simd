package result

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestAddition64NativeWidthAndComparisonContext(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkAddition64/", 1)
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "64", GeneratedInputs: true})
	if e != nil || run.HarnessID != LegacyAddition64HarnessID {
		t.Fatal(run, e)
	}
	o := run.Observations[0]
	if o.Key.Case.TypeID != "float64" || o.Key.Case.Vector.NativeLanes.Value != 8 || o.Key.Case.Input.SHA256 != input.Addition64Fingerprint(9) || len(o.ArrayResultSHA256) != 64 {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, "8 simd-lanes", "", 1), strings.Replace(raw, "8 simd-lanes", "1.5 simd-lanes", 1), strings.Replace(raw, "0 output-sha256-2", "", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt 64 metrics")
		}
	}
	historical, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "old"})
	if e != nil || historical.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("invented historical input")
	}
	a, b := o, o
	b.Label = "Scalar"
	b.Key.Case.StrategyID = "scalar"
	b.Key.Case.Vector.NativeLanes.Value = 4
	summary, e := Summarize([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Compare("mixed", summary); e == nil {
		t.Fatal("different widths pooled")
	}
	if _, e := NewVectorIdentity(0); e == nil {
		t.Fatal("zero width")
	}
}
