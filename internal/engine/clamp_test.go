package engine

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"math"
	"simd"
	"testing"
)

func TestClampPreparationAndFailureRetention(t *testing.T) {
	var v simd.Float32s
	options := input.DefaultOptions()
	options.Lower = 0
	options.Upper = 0x80000000
	p, err := PrepareWithOptions("clamp", Config{2, "5x", "^BenchmarkClamp$/^n=8$"}, v.Len(), options)
	if err != nil || len(p.Cases) != 3 {
		t.Fatal(p, err)
	}
	op, _ := catalog.ByGroup("clamp")
	x, _, out := input.MakeForOperation(op, 8, options)
	for i, value := range x {
		out[i] = value
		if value < 0 {
			out[i] = 0
		}
		if value > 0 {
			out[i] = math.Float32frombits(0x80000000)
		}
	}
	want := input.FingerprintFloat32(out, nil)
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: ClampHarnessID}
	run, err := Execute(p, r, &fakeBackend{})
	if err != nil || len(run.Observations) != 6 {
		t.Fatal(run, err)
	}
	for _, o := range run.Observations {
		if o.ArrayResultSHA256 != want || o.Key.Case.Clamp.PolicyID != result.ClampPolicyID || o.Key.Case.Clamp.NativeLanes.Value != v.Len() || o.Iterations != 5 {
			t.Fatal(o)
		}
	}
	partial, err := Execute(p, r, &fakeBackend{fail: 2})
	if err == nil || len(partial.Observations) != 1 || partial.Observations[0].ArrayResultSHA256 != want {
		t.Fatal(partial, err)
	}
}
