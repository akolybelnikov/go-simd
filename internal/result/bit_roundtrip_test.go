package result

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"strings"
	"testing"
)

func TestBitRoundTripNativeFacts(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkBitRoundTrip/", 1)
	run, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "bits", GeneratedInputs: true})
	if e != nil || run.HarnessID != LegacyBitRoundTripHarnessID {
		t.Fatal(run, e)
	}
	o := run.Observations[0]
	op, _ := catalog.ByGroup("bitroundtrip")
	if o.Key.Case.Vector.NativeLanes.Value != 8 || o.Key.Case.Input.SHA256 != input.FingerprintForOperation(op, 9, input.DefaultOptions()) || len(o.ArrayResultSHA256) != 64 || o.Key.Case.Copy != (CopyIdentity{}) {
		t.Fatal(o)
	}
	for _, bad := range []string{strings.Replace(raw, "8 simd-lanes", "", 1), strings.Replace(raw, "8 simd-lanes", "1.5 simd-lanes", 1), strings.Replace(raw, "0 output-sha256-2", "", 1)} {
		if _, e := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); e == nil {
			t.Fatal("corrupt roundtrip metrics")
		}
	}
	old, e := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "old"})
	if e != nil || old.Observations[0].Key.Case.Input.SHA256 != "" {
		t.Fatal("invented historical input")
	}
}
