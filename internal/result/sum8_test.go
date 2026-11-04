package result

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestSum8NativeExtentAndComparison(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkSum8/", 1)
	raw = strings.TrimSpace(raw) + " 2 output-elements\n"
	opts := input.DefaultOptions()
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "sum8", GeneratedInputs: true, InputOptions: &opts})
	if e != nil {
		t.Fatal(e)
	}
	o := run.Observations[0]
	if o.Key.Case.OutputElements != (OptionalInt{Value: 2, Known: true}) || len(o.ArrayResultSHA256) != 64 {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, "2 output-elements", "", 1), strings.Replace(raw, "2 output-elements", "3 output-elements", 1), strings.Replace(raw, "8 simd-lanes", "", 1), strings.Replace(raw, "0 output-sha256-2", "", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt projected evidence accepted")
		}
	}
	historical, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "old", GeneratedInputs: true})
	if e != nil || historical.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("invented historical recipe", e)
	}
	a, b := o, o
	b.Key.Case.Elements = 17
	b.Key.Case.OutputElements.Value = 3
	b.Key.Case.Input.SHA256 = ""
	a.Key.Case.Input.SHA256 = ""
	summaries, e := Summarize([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	table, e := Compare("sum8", summaries)
	if e != nil || table.BaselinePolicy != "specified-float32-block8-pair-fold/v1" {
		t.Fatal(table, e)
	}
	b = a
	b.Key.Case.StrategyID = catalog.Scalar
	b.Label = "Scalar"
	b.Key.Case.OutputElements.Value++
	summaries, e = Summarize([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Compare("bad", summaries); e == nil {
		t.Fatal("contradictory extent accepted")
	}
}
