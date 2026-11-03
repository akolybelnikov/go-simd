package result

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func predicateRaw(t *testing.T) string {
	t.Helper()
	digest, _ := hex.DecodeString("00010203ffffffff000000001234567889abcdef800000007fffffff00112233")
	text := "BenchmarkPredicate/n=9/SIMDPartial-4 3 12 ns/op 8 simd-lanes"
	for i := 0; i < 8; i++ {
		text += fmt.Sprintf(" %d output-sha256-%d", binary.BigEndian.Uint32(digest[i*4:]), i)
	}
	return text + "\n"
}

func TestPredicateLegacyFacts(t *testing.T) {
	raw := predicateRaw(t)
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "r"})
	if err != nil || len(run.Observations) != 1 || run.HarnessID != LegacyPredicateHarnessID {
		t.Fatal(run, err)
	}
	o := run.Observations[0]
	if o.ByteResultSHA256 != "00010203ffffffff000000001234567889abcdef800000007fffffff00112233" || o.Key.Case.Predicate.EncodingID != PredicateEncodingID || o.Key.Case.Predicate.NativeLanes.Value != 8 || o.ScalarResultBits != nil || o.ArrayResultLastBits != nil {
		t.Fatal(o)
	}
	for _, bad := range []string{
		strings.Replace(raw, " 8 simd-lanes", "", 1),
		strings.Replace(raw, " 8 simd-lanes", " 1.5 simd-lanes", 1),
		strings.Replace(raw, " 8 simd-lanes", " 0 simd-lanes", 1),
		strings.Replace(raw, "4294967295 output-sha256-1", "4294967296 output-sha256-1", 1),
		strings.Replace(raw, "0 output-sha256-2", "0.5 output-sha256-2", 1),
		strings.Replace(raw, "0 output-sha256-2", "-1 output-sha256-2", 1),
		strings.Replace(raw, "0 output-sha256-2", "NaN output-sha256-2", 1),
		strings.Replace(raw, "0 output-sha256-2", "", 1),
		strings.TrimSpace(raw) + " 0 output-sha256-2\n",
	} {
		if _, err := ReadLegacy(strings.NewReader(bad), LegacyOptions{RunID: "bad"}); err == nil {
			t.Fatal("accepted corrupt predicate facts", bad)
		}
	}
	if _, err := NewPredicateIdentity(0); err == nil {
		t.Fatal("invalid width")
	}
	for _, field := range []string{"width", "encoding"} {
		a, b := o, o
		b.Key.Case.StrategyID = "scalar"
		b.Label = "Scalar"
		if field == "width" {
			b.Key.Case.Predicate.NativeLanes.Value = 4
		} else {
			b.Key.Case.Predicate.EncodingID = "packed-bits/v1"
		}
		s, err := Summarize([]Observation{a, b})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Compare("mixed", s); err == nil {
			t.Fatal("mixed predicate context accepted", field)
		}
	}
}
