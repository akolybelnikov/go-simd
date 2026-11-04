package engine

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestSaturatingAdd8OutputAndIdentity(t *testing.T) {
	var f32 simd.Float32s
	var vector simd.Uint8s
	p, e := Prepare("saturatingadd8", Config{2, "5x", "^BenchmarkSaturatingAdd8$/^n=8$"}, f32.Len())
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	a, b, out := input.MakeSaturatingAdd8(8)
	for i := range out {
		out[i] = uint8(min(int(a[i])+int(b[i]), 255))
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: SaturatingAdd8HarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, o := range run.Observations {
		if o.ByteResultSHA256 != input.FingerprintUint8(out, nil) || o.Key.Case.Input.SHA256 != input.FingerprintUint8(a, b) || o.Key.Case.Vector.NativeLanes.Value != vector.Len() || o.Key.Case.Predicate != (result.PredicateIdentity{}) || o.Iterations != 5 {
			t.Fatal(o)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal(partial, e)
	}
}
