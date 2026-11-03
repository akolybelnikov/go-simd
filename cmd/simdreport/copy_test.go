package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestCopyResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("copy")
	sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 2*w + 1: true, 65536: true}
	selection := fmt.Sprintf("^BenchmarkCopy$/^n=(1|%d|%d|%d|%d|65536)$", w-1, w, w+1, 2*w+1)
	var runs []result.Run
	for _, backend := range []string{"compiled", "go-test"} {
		output := t.TempDir()
		stdout, stderr, err := runReportProcess(t, "-group=copy", "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output, "-pattern=shuffled", "-seed=123", "-alpha=NaN")
		if err != nil {
			t.Fatalf("%s: %v\n%s\n%s", backend, err, stdout, stderr)
		}
		files, _ := filepath.Glob(filepath.Join(output, "copy", "*"))
		if len(files) != 2 {
			t.Fatal(files)
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
				if run.HarnessID != engine.CopyHarnessID || c.Status != "complete" {
					t.Fatal(c)
				}
			} else if strings.HasSuffix(file, ".raw") {
				run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
				if err != nil || run.HarnessID != result.LegacyCopyHarnessID {
					t.Fatal(run, err)
				}
			}
		}
		if len(run.Observations) != len(sizes)*8 {
			t.Fatal("copy count", len(run.Observations))
		}
		labels := map[string]bool{}
		for _, o := range run.Observations {
			c := o.Key.Case
			want := input.FingerprintForOperation(op, c.Elements, input.DefaultOptions())
			labels[o.Label] = true
			if !sizes[c.Elements] || c.OperationID != catalog.Copy || c.ArithmeticID != op.ArithmeticID || c.Parameters != "" || c.Input.GeneratorID != input.CopyGeneratorID || c.Input.SHA256 != want || o.ArrayResultSHA256 != want || c.Copy.BitPolicyID != result.CopyBitPolicyID || c.Copy.NativeLanes.Value != w || o.Iterations != 3 || o.ByteResultSHA256 != "" || o.ArrayResultLastBits != nil || o.ScalarResultBits != nil {
				t.Fatal("copy evidence/context lost", o)
			}
		}
		for _, label := range []string{"Scalar", "GoCopy", "SIMDPartial", "SIMDTail"} {
			if !labels[label] {
				t.Fatal("missing copy control", label)
			}
		}
		runs = append(runs, run)
	}
	for i, a := range runs[0].Observations {
		b := runs[1].Observations[i]
		if a.Key.Case != b.Key.Case || a.ArrayResultSHA256 != b.ArrayResultSHA256 {
			t.Fatal("copy backend parity")
		}
	}
}
