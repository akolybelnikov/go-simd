package engine

import (
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestMulAddOutputAndPartialRetention(t *testing.T) {
	var v simd.Float32s
	o := input.DefaultOptions()
	o.FMAInput = "rounding"
	p, e := PrepareWithOptions("muladd", Config{2, "3x", "^BenchmarkMulAdd$/^n=8$"}, v.Len(), o)
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	a, b, c, out := input.MakeMulAdd(8, o)
	for i := range out {
		out[i] = fma32.FMA(a[i], b[i], c[i])
	}
	want := input.FingerprintFloat32(out, nil)
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: MulAddHarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, sample := range run.Observations {
		if sample.ArrayResultSHA256 != want || sample.Key.Case.Input.SHA256 != input.FingerprintTernaryFloat32(a, b, c) || sample.Key.Case.FMA.RoundingID != fma32.RoundingID || sample.Key.Case.FMA.VectorPathID != fmacap.Path() || sample.Key.Case.FMA.NativeLanes.Value != v.Len() || sample.Iterations != 5 {
			t.Fatal(sample)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 || partial.Observations[0].ArrayResultSHA256 != want {
		t.Fatal(partial, e)
	}
}
