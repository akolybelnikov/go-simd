package engine

import (
	"crypto/sha256"
	"fmt"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestPackedPredicatePreparedAndPartialFailure(t *testing.T) {
	var v simd.Float32s
	p, err := Prepare("predicatepack", Config{2, "5x", "^BenchmarkPredicatePacked$/^n=(8|9)$"}, v.Len())
	if err != nil || len(p.Cases) != 4 {
		t.Fatal(err, p)
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: PackedPredicateHarnessID}
	run, err := Execute(p, r, &fakeBackend{})
	if err != nil || len(run.Observations) != 8 {
		t.Fatal(run, err)
	}
	for _, o := range run.Observations {
		out := []byte{0x55}
		if o.Key.Case.Elements == 9 {
			out = append(out, 1)
		}
		digest := sha256.Sum256(out)
		if o.ByteResultSHA256 != fmt.Sprintf("%x", digest) || o.Key.Case.Predicate.EncodingID != result.PackedPredicateEncodingID || o.Key.Case.Predicate.NativeLanes.Value != v.Len() || o.Iterations != 5 {
			t.Fatal(o)
		}
	}
	partial, err := Execute(p, r, &fakeBackend{fail: 2})
	if err == nil || len(partial.Observations) != 1 || len(partial.Observations[0].ByteResultSHA256) != 64 {
		t.Fatal("partial packed capture lost", err)
	}
}
