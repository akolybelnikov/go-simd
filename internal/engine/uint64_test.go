package engine

import (
	"errors"
	"flag"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"github.com/akolybelnikov/go-simd/internal/result"
	"simd"
	"testing"
)

func TestClmul64OutputAndActualWidth(t *testing.T) {
	var f32 simd.Float32s
	var f64 simd.Uint64s
	p, e := Prepare("clmul64", Config{2, "5x", "^BenchmarkClmul64$/^n=8$"}, f32.Len())
	if e != nil || len(p.Cases) != 2 {
		t.Fatal(p, e)
	}
	a, b, out := input.MakeClmul64(8)
	if e := integer.Clmul64Scalar(a, b, out); e != nil {
		t.Fatal(e)
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: Clmul64HarnessID}
	run, e := Execute(p, r, &fakeBackend{})
	if e != nil || len(run.Observations) != 4 {
		t.Fatal(run, e)
	}
	for _, sample := range run.Observations {
		if sample.ArrayResultSHA256 != input.FingerprintUint64(out, nil) || sample.Key.Case.Input.SHA256 != input.FingerprintUint64(a, b) || sample.Key.Case.Vector.NativeLanes.Value != f64.Len() || sample.Iterations != 5 || sample.Key.Case.Vector.CarrylessCapabilityID != result.CarrylessCapability(simd.HasHardwareCarrylessMultiply()) {
			t.Fatal(sample)
		}
	}
	partial, e := Execute(p, r, &fakeBackend{fail: 2})
	if e == nil || len(partial.Observations) != 1 {
		t.Fatal(partial, e)
	}
}
func TestUint64AdapterRetentionAndFlagRestoration(t *testing.T) {
	out := []uint64{0}
	prior := flag.Lookup("test.benchtime").Value.String()
	p := Prepared{Run: func() error { out[0] = 1; return nil }, Uint64Output: out, Bytes: 24, RetainOutputAfterTiming: true}
	sample, e := (Testing{}).Measure(p, "3x")
	if e != nil || sample.Iterations != 3 || retainedUint64Sink[0] != 1 {
		t.Fatal(sample, e)
	}
	want := errors.New("word failure")
	p.Run = func() error { return want }
	if _, e := (Testing{}).Measure(p, "1x"); !errors.Is(e, want) || flag.Lookup("test.benchtime").Value.String() != prior {
		t.Fatal(e)
	}
	p.Output = []float32{0}
	if _, e := (Testing{}).Measure(p, "1x"); e == nil {
		t.Fatal("mixed output shape")
	}
}
