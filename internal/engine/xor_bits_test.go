package engine

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"math"
	"simd"
	"testing"
)

func TestXorBitsCompleteOutputAndWidth(t *testing.T) {
	var vector simd.Float32s
	p, e := Prepare("xorbits", Config{2, "5x", "^BenchmarkXorBits$/^n=8$"}, vector.Len())
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	a, b, out := input.MakeXorBits(8)
	for i := range out {
		out[i] = math.Float32frombits(math.Float32bits(a[i]) ^ math.Float32bits(b[i]))
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: XorBitsHarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, o := range run.Observations {
		if o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || o.Key.Case.Input.SHA256 != input.FingerprintFloat32(a, b) || o.Key.Case.Vector.NativeLanes.Value != vector.Len() {
			t.Fatal(o)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal(partial, e)
	}
}
