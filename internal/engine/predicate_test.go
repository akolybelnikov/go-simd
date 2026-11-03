package engine

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestPredicatePreparedAndFailureRetention(t *testing.T) {
	var v simd.Float32s
	p, err := Prepare("predicate", Config{2, "5x", "^BenchmarkPredicate$/^n=8$"}, v.Len())
	if err != nil || len(p.Cases) != 2 {
		t.Fatal(err)
	}
	r := result.Run{ID: "r", Environment: result.Environment{ID: "e"}, BackendID: BackendID, HarnessID: PredicateHarnessID}
	run, err := Execute(p, r, &fakeBackend{})
	want := sha256.Sum256([]byte{1, 0, 1, 0, 1, 0, 1, 0})
	if err != nil || len(run.Observations) != 4 {
		t.Fatal(err, run)
	}
	for _, o := range run.Observations {
		if o.ByteResultSHA256 != fmt.Sprintf("%x", want) || o.Key.Case.Predicate.NativeLanes.Value != v.Len() || o.Key.Case.Predicate.EncodingID != result.PredicateEncodingID || o.Iterations != 5 || o.ScalarResultBits != nil {
			t.Fatal(o)
		}
	}
	partial, err := Execute(p, r, &fakeBackend{fail: 2})
	if err == nil || len(partial.Observations) != 1 || partial.Observations[0].ByteResultSHA256 != fmt.Sprintf("%x", want) {
		t.Fatal("partial mask capture lost", err)
	}
}

func TestPredicateTestingAdapter(t *testing.T) {
	previous := flag.Lookup("test.benchtime").Value.String()
	out := []byte{7}
	s, err := (Testing{}).Measure(Prepared{Run: func() error { out[0] = 1; return nil }, ByteOutput: out, Bytes: 5}, "3x")
	if err != nil || s.Iterations != 3 || out[0] != 1 {
		t.Fatal(s, err)
	}
	want := errors.New("byte kernel failure")
	_, err = (Testing{}).Measure(Prepared{Run: func() error { return want }, ByteOutput: out, Bytes: 5}, "2x")
	if !errors.Is(err, want) || flag.Lookup("test.benchtime").Value.String() != previous {
		t.Fatal(err)
	}
	for _, p := range []Prepared{
		{ByteOutput: out},
		{Run: func() error { return nil }, ByteOutput: []byte{}},
		{Run: func() error { return nil }, ByteOutput: out, Output: []float32{0}},
		{Run: func() error { return nil }, ByteOutput: out, Reduce: func() float32 { return 0 }},
	} {
		if _, err := (Testing{}).Measure(p, "1x"); err == nil {
			t.Fatal("mixed/invalid output shape accepted")
		}
	}
}
