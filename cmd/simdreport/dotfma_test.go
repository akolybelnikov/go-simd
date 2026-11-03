package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestDotFMAResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("dotfma")
	sizes := map[int]bool{1: true, w: true, w + 1: true, 64: true}
	selection := fmt.Sprintf("^BenchmarkDotFMA$/^n=(1|%d|%d|64)$", w, w+1)
	for _, fixture := range []string{"balanced", "cancellation", "mixed-magnitude", "fusion-sensitive"} {
		o := input.DefaultOptions()
		o.DotInput = fixture
		var runs []result.Run
		for _, backend := range []string{"compiled", "go-test"} {
			output := t.TempDir()
			stdout, stderr, err := runReportProcess(t, "-group=dotfma", "-dot-input="+fixture, "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output)
			if err != nil {
				t.Fatalf("%s/%s: %v\n%s\n%s", fixture, backend, err, stdout, stderr)
			}
			files, _ := filepath.Glob(filepath.Join(output, "dotfma", "*"))
			if len(files) != 2 {
				t.Fatal("sibling selection/artifacts", files)
			}
			var run result.Run
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, ".measurements.json") {
					var c evidence.Capture
					if err := json.Unmarshal(data, &c); err != nil {
						t.Fatal(err)
					}
					run = c.Run
					if run.HarnessID != engine.DotFMAHarnessID || c.Machine.FMA32Path != fmacap.Path() {
						t.Fatal("compiled fused-dot metadata")
					}
				}
				if strings.HasSuffix(file, ".raw") {
					run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
					if err != nil || run.HarnessID != result.LegacyDotFMAHarnessID {
						t.Fatal("native fused-dot metadata", err)
					}
				}
			}
			if len(run.Observations) != len(sizes)*8 {
				t.Fatal("wrong fused-dot sample count", len(run.Observations))
			}
			for _, obs := range run.Observations {
				c := obs.Key.Case
				x, y, _ := input.MakeForOperation(op, c.Elements, o)
				kernel := reduction.DotFMAScalar
				switch obs.Label {
				case "SIMDStriped":
					kernel = reduction.DotFMASIMD
				case "SIMDStriped2":
					kernel = reduction.DotFMASIMD2
				case "SIMDStriped4":
					kernel = reduction.DotFMASIMD4
				}
				want, err := kernel(x, y)
				if err != nil {
					t.Fatal(err)
				}
				identity, err := result.NewDotFMAIdentity(c.StrategyID, fmacap.Path(), w, c.Elements)
				if err != nil {
					t.Fatal(err)
				}
				if c.OperationID != catalog.DotFMA || !sizes[c.Elements] || c.FMA != identity || c.Input.SHA256 != input.FingerprintFloat32(x, y) || obs.ScalarResultBits == nil || *obs.ScalarResultBits != math.Float32bits(want) || obs.Iterations != 3 {
					t.Fatal("fused-dot result/context lost", obs)
				}
			}
			runs = append(runs, run)
		}
		for i, a := range runs[0].Observations {
			b := runs[1].Observations[i]
			if a.Key.Case != b.Key.Case || *a.ScalarResultBits != *b.ScalarResultBits {
				t.Fatal("fused-dot backend parity")
			}
		}
	}
}
