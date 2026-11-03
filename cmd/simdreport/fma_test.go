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
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestFMAResultsAndBackendIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("affinefma")
	for _, fixture := range []string{"balanced", "rounding", "mixed-magnitude"} {
		o := input.DefaultOptions()
		o.FMAInput = fixture
		o.Alpha = math.Float32bits(1 + 0x1p-23)
		o.Beta = math.Float32bits(-1)
		var runs []result.Run
		for _, backend := range []string{"compiled", "go-test"} {
			output := t.TempDir()
			stdout, stderr, err := runReportProcess(t, "-group=affinefma", "-fma-input="+fixture, "-alpha=1.0000001192092896", "-beta=-1", "-backend="+backend, "-bench=^BenchmarkAffineFMA$/^n=(8|16)$", "-count=2", "-benchtime=3x", "-format=md", "-output", output)
			if err != nil {
				t.Fatalf("%s/%s: %v\n%s\n%s", fixture, backend, err, stdout, stderr)
			}
			files, _ := filepath.Glob(filepath.Join(output, "affinefma", "*"))
			if len(files) != 2 {
				t.Fatal("wrong FMA artifacts", files)
			}
			var run result.Run
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, ".measurements.json") {
					var capture evidence.Capture
					if err := json.Unmarshal(data, &capture); err != nil {
						t.Fatal(err)
					}
					run = capture.Run
					if run.HarnessID != engine.FMAHarnessID || capture.Machine.FMA32Path != patterns.FMAPath() || capture.Config.Input.FMAInput != fixture {
						t.Fatal("compiled FMA context lost")
					}
				}
				if strings.HasSuffix(file, ".raw") {
					run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
					if err != nil || run.HarnessID != result.LegacyFMAHarnessID {
						t.Fatal("native FMA context", err)
					}
				}
			}
			if len(run.Observations) != 8 {
				t.Fatal("FMA sibling selection or repetition drift")
			}
			for _, obs := range run.Observations {
				c := obs.Key.Case
				x, _, _ := input.MakeForOperation(op, c.Elements, o)
				want := math.Float32bits(fma32.FMA(o.Float("alpha"), x[len(x)-1], o.Float("beta")))
				if c.OperationID != catalog.AffineFMA || c.Parameters != o.Identity(op) || c.Input.SHA256 != input.FingerprintFloat32(x, nil) || c.FMA.VectorPathID != patterns.FMAPath() || obs.Iterations != 3 || obs.ArrayResultLastBits == nil || *obs.ArrayResultLastBits != want {
					t.Fatal("FMA numerical/context facts lost", obs)
				}
				if c.StrategyID == catalog.Scalar && c.FMA.PathID != fma32.SoftwarePath {
					t.Fatal("scalar reference mislabeled")
				}
			}
			runs = append(runs, run)
		}
		for i, a := range runs[0].Observations {
			b := runs[1].Observations[i]
			if a.Key.Case != b.Key.Case || *a.ArrayResultLastBits != *b.ArrayResultLastBits {
				t.Fatal("native/compiled FMA mismatch")
			}
		}
	}
	for _, args := range [][]string{{"-group=affinefma", "-fma-input=bad"}, {"-group=affinefma", "-alpha=NaN", "-beta=Inf", "-bench=^BenchmarkAffineFMA$/^n=8$", "-count=1", "-benchtime=1x"}} {
		output := t.TempDir()
		args = append(args, "-output", output)
		_, stderr, err := runReportProcess(t, args...)
		if strings.Contains(strings.Join(args, " "), "fma-input=bad") {
			if err == nil {
				t.Fatal("invalid FMA fixture accepted")
			}
			if files, _ := os.ReadDir(output); len(files) != 0 {
				t.Fatal("invalid FMA fixture created artifacts")
			}
		} else if err != nil {
			t.Fatal("nonfinite FMA bits broke JSON", err, stderr)
		}
	}
}
