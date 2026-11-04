package engine

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestCopy64TypedCompleteEvidence(t *testing.T) {
	var vector simd.Float32s
	var actual simd.Float64s
	p, e := Prepare("copy64", Config{2, "5x", "^BenchmarkCopy64$/^n=8$"}, vector.Len())
	if e != nil || len(p.Cases) != 3 {
		t.Fatal(p, e)
	}
	x, _ := input.MakeCopy64(8)
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: Copy64HarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 6 {
		t.Fatal(e)
	}
	for _, o := range run.Observations {
		if o.ArrayResultSHA256 != input.FingerprintFloat64(x, nil) || o.Key.Case.Input.SHA256 != o.ArrayResultSHA256 || o.Key.Case.Vector.NativeLanes.Value != actual.Len() {
			t.Fatal(o)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal("partial", e)
	}
}
