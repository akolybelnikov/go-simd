package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestReductionBackendResultAndIdentityParity(t *testing.T) {
	var runs []result.Run
	for _, backend := range []string{"compiled", "go-test"} {
		output := t.TempDir()
		stdout, stderr, err := runReportProcess(t, "-group=sum", "-sum-input=cancellation", "-backend="+backend, "-bench=^BenchmarkSum$/^n=64$", "-count=2", "-benchtime=3x", "-format=md", "-output", output)
		if err != nil {
			t.Fatalf("%s: %v\n%s\n%s", backend, err, stdout, stderr)
		}
		paths, _ := filepath.Glob(filepath.Join(output, "sum", "*"))
		if len(paths) != 2 {
			t.Fatal(paths)
		}
		var run result.Run
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasSuffix(path, ".measurements.json") {
				var capture evidence.Capture
				if err := json.Unmarshal(data, &capture); err != nil {
					t.Fatal(err)
				}
				run = capture.Run
				if run.HarnessID != engine.ReductionHarnessID {
					t.Fatal("wrong direct harness")
				}
			}
			if strings.HasSuffix(path, ".raw") {
				run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
				if err != nil {
					t.Fatal(err)
				}
				if run.HarnessID != result.LegacyReductionHarnessID {
					t.Fatal("wrong native harness")
				}
			}
		}
		if len(run.Observations) != 8 {
			t.Fatal("missing sum observations")
		}
		for _, o := range run.Observations {
			if o.ScalarResultBits == nil || o.Key.Case.Reduction.OrderID == "" || o.Key.Case.Input.SHA256 == "" {
				t.Fatal("sum evidence incomplete")
			}
			want := uint32(0)
			if o.Label != "Scalar" {
				want = math.Float32bits(1)
			}
			if *o.ScalarResultBits != want {
				t.Fatalf("wrong numerical result for %s", o.Label)
			}
		}
		runs = append(runs, run)
	}
	for i := range runs[0].Observations {
		a, b := runs[0].Observations[i], runs[1].Observations[i]
		if a.Key.Case != b.Key.Case || *a.ScalarResultBits != *b.ScalarResultBits {
			t.Fatal("compiled/native sum context mismatch")
		}
	}
}
