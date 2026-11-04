package engine

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestSum8ProjectedArrayCapture(t *testing.T) {
	var vector simd.Float32s
	p, e := Prepare("sum8", Config{2, "5x", "^BenchmarkSum8$/^n=9$"}, vector.Len())
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	op, _ := catalog.ByGroup("sum8")
	x, _, out := input.MakeForOperation(op, 9, input.DefaultOptions())
	if e := reduction.Sum8Scalar(x, out); e != nil {
		t.Fatal(e)
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: Sum8HarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, o := range run.Observations {
		if !o.Key.Case.OutputElements.Known || o.Key.Case.OutputElements.Value != 2 || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || o.Key.Case.Input.SHA256 != input.FingerprintFloat32(x, nil) || o.ScalarResultBits != nil {
			t.Fatal(o)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal(partial, e)
	}
}
