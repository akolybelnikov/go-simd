package engine

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestBitRoundTripCompleteOutput(t *testing.T) {
	var vector simd.Float32s
	p, e := Prepare("bitroundtrip", Config{2, "5x", "^BenchmarkBitRoundTrip$/^n=8$"}, vector.Len())
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	op, _ := catalog.ByGroup("bitroundtrip")
	want := input.FingerprintForOperation(op, 8, input.DefaultOptions())
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: BitRoundTripHarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, o := range run.Observations {
		if o.ArrayResultSHA256 != want || o.Key.Case.Input.SHA256 != want || o.Key.Case.Vector.NativeLanes.Value != vector.Len() || o.Key.Case.Copy != (result.CopyIdentity{}) {
			t.Fatal(o)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal(partial, e)
	}
}
