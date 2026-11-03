package engine

import (
	"errors"
	"flag"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestCopyPreparationAndFailureRetention(t *testing.T) {
	var v simd.Float32s
	p, err := Prepare("copy", Config{2, "5x", "^BenchmarkCopy$/^n=8$"}, v.Len())
	if err != nil || len(p.Cases) != 4 {
		t.Fatal(p, err)
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: CopyHarnessID}
	run, err := Execute(p, r, &fakeBackend{})
	if err != nil || len(run.Observations) != 8 {
		t.Fatal(run, err)
	}
	op, _ := catalog.ByGroup("copy")
	want := input.FingerprintForOperation(op, 8, input.DefaultOptions())
	for _, o := range run.Observations {
		if o.ArrayResultSHA256 != want || o.Key.Case.Input.SHA256 != want || o.Key.Case.Copy.BitPolicyID != result.CopyBitPolicyID || o.Key.Case.Copy.NativeLanes.Value != v.Len() || o.Iterations != 5 || o.ByteResultSHA256 != "" || o.ScalarResultBits != nil {
			t.Fatal(o)
		}
	}
	partial, err := Execute(p, r, &fakeBackend{fail: 2})
	if err == nil || len(partial.Observations) != 1 || partial.Observations[0].ArrayResultSHA256 != want {
		t.Fatal(partial, err)
	}
}

func TestCopyRetainedArrayAdapter(t *testing.T) {
	previous := flag.Lookup("test.benchtime").Value.String()
	out := []float32{0}
	s, err := (Testing{}).Measure(Prepared{Run: func() error { out[0] = 1; return nil }, Output: out, Bytes: 8, RetainOutputAfterTiming: true}, "3x")
	if err != nil || s.Iterations != 3 || out[0] != 1 || retainedArraySink[0] != 1 {
		t.Fatal(s, err)
	}
	want := errors.New("copy failure")
	_, err = (Testing{}).Measure(Prepared{Run: func() error { return want }, Output: out, Bytes: 8, RetainOutputAfterTiming: true}, "2x")
	if !errors.Is(err, want) || flag.Lookup("test.benchtime").Value.String() != previous {
		t.Fatal(err)
	}
	for _, p := range []Prepared{{RetainOutputAfterTiming: true}, {RetainOutputAfterTiming: true, Run: func() error { return nil }, ByteOutput: []byte{0}}, {RetainOutputAfterTiming: true, Reduce: func() float32 { return 0 }, Scalar: new(float32)}} {
		if _, err := (Testing{}).Measure(p, "1x"); err == nil {
			t.Fatal("invalid copy shape")
		}
	}
}
