package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestDotBackendResultAndIdentityParity(t *testing.T) {
	op, _ := catalog.ByGroup("dot")
	for _, fixture := range []string{"balanced", "cancellation", "mixed-magnitude"} {
		options := input.DefaultOptions()
		options.DotInput = fixture
		x, y, _ := input.MakeForOperation(op, 64, options)
		scalar, _ := reduction.DotScalar(x, y)
		striped, _ := reduction.DotSIMD(x, y)
		var runs []result.Run
		for _, backend := range []string{"compiled", "go-test"} {
			output := t.TempDir()
			stdout, stderr, err := runReportProcess(t, "-group=dot", "-dot-input="+fixture, "-backend="+backend, "-bench=^BenchmarkDot$/^n=64$", "-count=2", "-benchtime=3x", "-format=md", "-output", output)
			if err != nil {
				t.Fatalf("%s/%s: %v\n%s\n%s", fixture, backend, err, stdout, stderr)
			}
			paths, _ := filepath.Glob(filepath.Join(output, "dot", "*"))
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
					if run.HarnessID != engine.DotHarnessID || capture.Config.Input.DotInput != fixture || !strings.Contains(run.Metadata, "logical traffic: 8*n + 4") {
						t.Fatal("dot compiled context lost")
					}
				}
				if strings.HasSuffix(path, ".raw") {
					run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
					if err != nil || run.HarnessID != result.LegacyDotHarnessID {
						t.Fatal("dot native context lost", err)
					}
				}
			}
			if len(run.Observations) != 4 {
				t.Fatal("missing dot observations")
			}
			for _, o := range run.Observations {
				want := math.Float32bits(scalar)
				if o.Label == "SIMDStriped" {
					want = math.Float32bits(striped)
				}
				if o.Key.Case.OperationID != catalog.Dot || o.Key.Case.Elements != 64 || o.ScalarResultBits == nil || *o.ScalarResultBits != want || o.Key.Case.Input.SHA256 != input.FingerprintFloat32(x, y) || o.Iterations != 3 {
					t.Fatal("wrong dot result/selection/context", o)
				}
			}
			runs = append(runs, run)
		}
		for i, a := range runs[0].Observations {
			b := runs[1].Observations[i]
			if a.Key.Case != b.Key.Case || *a.ScalarResultBits != *b.ScalarResultBits {
				t.Fatal("compiled/native dot mismatch")
			}
		}
	}
	output := t.TempDir()
	if _, _, err := runReportProcess(t, "-group=dot", "-dot-input=unknown", "-output", output); err == nil {
		t.Fatal("invalid dot fixture accepted")
	}
	if files, _ := os.ReadDir(output); len(files) != 0 {
		t.Fatal("invalid dot options created artifacts")
	}
}
