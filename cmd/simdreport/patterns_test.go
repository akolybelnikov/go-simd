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
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestPatternBackendIdentityParity(t *testing.T) {
	for _, group := range []string{"scale", "affine", "threshold", "selection", "rangefilter"} {
		t.Run(group, func(t *testing.T) {
			op, _ := catalog.ByGroup(group)
			var captures []result.Run
			for _, backend := range []string{"compiled", "go-test"} {
				output := t.TempDir()
				args := []string{"-group=" + group, "-backend=" + backend, "-bench=^" + op.BenchmarkName + "$/^n=(8|16)$", "-count=2", "-benchtime=2x", "-alpha=-0", "-beta=0.25", "-threshold=-1", "-pattern=shuffled", "-seed=123", "-output", output, "-format=csv"}
				stdout, stderr, err := runReportProcess(t, args...)
				if err != nil {
					t.Fatalf("%s: %v\n%s\n%s", backend, err, stdout, stderr)
				}
				files, _ := filepath.Glob(filepath.Join(output, group, "*"))
				if len(files) != 2 {
					t.Fatalf("sibling operation measured: %v", files)
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
						if capture.Config.Input == nil || capture.Config.Input.Alpha != 0x80000000 || capture.Config.Input.Seed != 123 || capture.Run.HarnessID != engine.PatternHarnessID {
							t.Fatal("direct context lost")
						}
					} else if strings.HasSuffix(file, ".raw") {
						run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
						if err != nil {
							t.Fatal(err)
						}
						if run.HarnessID != result.LegacyPatternHarnessID {
							t.Fatal("native timing context lost")
						}
					}
				}
				if len(run.Observations) != 8 {
					t.Fatalf("wrong observation count: %d", len(run.Observations))
				}
				for _, observation := range run.Observations {
					if observation.Key.Case.OperationID != op.ID || observation.Key.Case.Parameters == "" || observation.Key.Case.Input.SHA256 == "" {
						t.Fatalf("identity missing: %+v", observation)
					}
				}
				captures = append(captures, run)
			}
			for i := range captures[0].Observations {
				if captures[0].Observations[i].Key.Case != captures[1].Observations[i].Key.Case {
					t.Fatalf("input/parameter/worker drift: %+v vs %+v", captures[0].Observations[i].Key.Case, captures[1].Observations[i].Key.Case)
				}
			}
		})
	}
}

func TestNonfiniteArithmeticParametersRemainReportable(t *testing.T) {
	output := t.TempDir()
	_, stderr, err := runReportProcess(t, "-group=affine", "-alpha=NaN", "-beta=+Inf", "-bench=^BenchmarkAffine$/^n=8$", "-count=1", "-benchtime=1x", "-output", output)
	if err != nil {
		t.Fatalf("nonfinite output prevented report: %v %s", err, stderr)
	}
	files, _ := filepath.Glob(filepath.Join(output, "affine", "*.measurements.json"))
	if len(files) != 1 {
		t.Fatal(files)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var capture evidence.Capture
	if err := json.Unmarshal(data, &capture); err != nil || capture.Status != "complete" || capture.Config.Input == nil || !math.IsNaN(float64(capture.Config.Input.Float("alpha"))) || !math.IsInf(float64(capture.Config.Input.Float("beta")), 1) {
		t.Fatalf("invalid nonfinite capture: %v %s", err, data)
	}
}

func TestPatternValidationBeforeArtifacts(t *testing.T) {
	for _, args := range [][]string{{"-pattern=invalid"}, {"-alpha=1e100"}, {"-group=threshold", "-threshold=+Inf"}, {"-group=rangefilter", "-lower=1", "-upper=-1"}} {
		output := filepath.Join(t.TempDir(), "reports")
		var stdout bytes.Buffer
		if err := run(append(args, "-output", output), &stdout); err == nil {
			t.Fatalf("invalid input accepted: %v", args)
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatal("invalid options created artifacts")
		}
	}
	op, _ := catalog.ByGroup("threshold")
	options := input.DefaultOptions()
	options.Threshold = 0x7f800000
	options.Pattern = "none"
	if _, err := engine.PrepareWithOptions(op.Group, engine.Config{Count: 1, Benchtime: "1x", Bench: "."}, 8, options); err != nil {
		t.Fatal(err)
	}
}
