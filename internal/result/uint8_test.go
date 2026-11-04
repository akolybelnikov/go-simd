package result

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestSaturatingAdd8NativeFacts(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkSaturatingAdd8/", 1)
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "u8", GeneratedInputs: true})
	if e != nil || run.HarnessID != LegacySaturatingAdd8HarnessID {
		t.Fatal(run, e)
	}
	o := run.Observations[0]
	if o.Key.Case.TypeID != "uint8" || o.Key.Case.Input.SHA256 != input.SaturatingAdd8Fingerprint(9) || o.Key.Case.Vector.NativeLanes.Value != 8 || o.Key.Case.Predicate != (PredicateIdentity{}) || len(o.ByteResultSHA256) != 64 {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, "8 simd-lanes", "", 1), strings.Replace(raw, "8 simd-lanes", "1.5 simd-lanes", 1), strings.Replace(raw, "0 output-sha256-2", "", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt byte metrics")
		}
	}
	old, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "old"})
	if e != nil || old.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("historical recipe invented")
	}
}
