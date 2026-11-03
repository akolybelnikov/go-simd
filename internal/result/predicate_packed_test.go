package result

import (
	"strings"
	"testing"
)

func TestPackedPredicateLegacyIdentity(t *testing.T) {
	raw := strings.Replace(predicateRaw(t), "BenchmarkPredicate/", "BenchmarkPredicatePacked/", 1)
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "packed"})
	if err != nil || run.HarnessID != LegacyPackedPredicateHarnessID || len(run.Observations) != 1 {
		t.Fatal(run, err)
	}
	identity, err := NewPackedPredicateIdentity(8)
	if err != nil || identity.EncodingID != PackedPredicateEncodingID || run.Observations[0].Key.Case.Predicate != identity || len(run.Observations[0].ByteResultSHA256) != 64 {
		t.Fatal(identity, err)
	}
	if _, err := NewPackedPredicateIdentity(0); err == nil {
		t.Fatal("invalid packed width")
	}
	plain, _ := NewPredicateIdentity(8)
	if plain.EncodingID == identity.EncodingID {
		t.Fatal("representations collapsed")
	}
}
