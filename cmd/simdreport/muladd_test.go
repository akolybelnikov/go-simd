package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"
)

func TestMulAddResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("muladd")
	for _, fixture := range []string{"balanced", "rounding", "mixed-magnitude"} {
		t.Run(fixture, func(t *testing.T) {
			options := input.DefaultOptions()
			options.FMAInput = fixture
			var runs []result.Run
			sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 64: true}
			selection := fmt.Sprintf("^BenchmarkMulAdd$/^n=(1|%d|%d|%d|64)$", w-1, w, w+1)
			for _, backend := range []string{"compiled", "go-test"} {
				output := t.TempDir()
				stdout, stderr, e := runReportProcess(t, "-group=muladd", "-fma-input="+fixture, "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-alpha=NaN", "-output", output)
				if e != nil {
					t.Fatalf("%s %v\n%s\n%s", backend, e, stdout, stderr)
				}
				files, _ := filepath.Glob(filepath.Join(output, "muladd", "*"))
				if len(files) != 2 {
					t.Fatal(files)
				}
				var run result.Run
				for _, file := range files {
					data, e := os.ReadFile(file)
					if e != nil {
						t.Fatal(e)
					}
					if strings.HasSuffix(file, ".measurements.json") {
						var capture evidence.Capture
						if e := json.Unmarshal(data, &capture); e != nil {
							t.Fatal(e)
						}
						run = capture.Run
						if capture.Status != "complete" || run.HarnessID != engine.MulAddHarnessID || capture.Config.Input.FMAInput != fixture {
							t.Fatal(capture)
						}
					} else if strings.HasSuffix(file, ".raw") {
						run, e = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
						if e != nil || run.HarnessID != result.LegacyMulAddHarnessID {
							t.Fatal(run, e)
						}
					}
				}
				if len(run.Observations) != len(sizes)*4 {
					t.Fatal(len(run.Observations))
				}
				for _, o := range run.Observations {
					c := o.Key.Case
					a, b, third, out := input.MakeMulAdd(c.Elements, options)
					for i := range out {
						out[i] = fma32.FMA(a[i], b[i], third[i])
					}
					identity, e := result.NewFMAIdentity(c.StrategyID, fmacap.Path(), w)
					if e != nil || !sizes[c.Elements] || c.OperationID != catalog.MulAdd || c.ArithmeticID != op.ArithmeticID || c.Parameters != "" || c.Input.GeneratorID != options.Generator(op) || c.Input.SHA256 != input.FingerprintTernaryFloat32(a, b, third) || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || c.FMA != identity || o.Iterations != 3 || o.ArrayResultLastBits != nil || o.ScalarResultBits != nil {
						t.Fatal("ternary evidence lost", o, e)
					}
				}
				runs = append(runs, run)
			}
			for i, a := range runs[0].Observations {
				b := runs[1].Observations[i]
				if a.Key.Case != b.Key.Case || a.ArrayResultSHA256 != b.ArrayResultSHA256 {
					t.Fatal("muladd backend drift", a, b)
				}
			}
		})
	}
}
