package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"math"
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"
)

func TestXorBitsResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 1024: true}
	selection := fmt.Sprintf("^BenchmarkXorBits$/^n=(1|%d|%d|%d|1024)$", w-1, w, w+1)
	var runs []result.Run
	for _, backend := range []string{"compiled", "go-test"} {
		output := t.TempDir()
		stdout, stderr, e := runReportProcess(t, "-group=xorbits", "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output)
		if e != nil {
			t.Fatalf("%s %v\n%s\n%s", backend, e, stdout, stderr)
		}
		files, _ := filepath.Glob(filepath.Join(output, "xorbits", "*"))
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
				var c evidence.Capture
				if e := json.Unmarshal(data, &c); e != nil {
					t.Fatal(e)
				}
				run = c.Run
				if c.Status != "complete" || run.HarnessID != engine.XorBitsHarnessID {
					t.Fatal(c)
				}
			} else if strings.HasSuffix(file, ".raw") {
				run, e = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
				if e != nil || run.HarnessID != result.LegacyXorBitsHarnessID {
					t.Fatal(run, e)
				}
			}
		}
		if len(run.Observations) != len(sizes)*4 {
			t.Fatal(len(run.Observations))
		}
		for _, o := range run.Observations {
			c := o.Key.Case
			a, b, out := input.MakeXorBits(c.Elements)
			for i := range out {
				out[i] = math.Float32frombits(math.Float32bits(a[i]) ^ math.Float32bits(b[i]))
			}
			if !sizes[c.Elements] || c.OperationID != catalog.XorBits || c.TypeID != catalog.Float32 || c.Parameters != "" || c.Input.GeneratorID != input.XorBitsGeneratorID || c.Input.SHA256 != input.FingerprintFloat32(a, b) || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || c.Vector.NativeLanes.Value != w || o.Iterations != 3 {
				t.Fatal("XOR evidence lost", o)
			}
		}
		runs = append(runs, run)
	}
	for i, a := range runs[0].Observations {
		b := runs[1].Observations[i]
		if a.Key.Case != b.Key.Case || a.ArrayResultSHA256 != b.ArrayResultSHA256 {
			t.Fatal("XOR backend drift", a, b)
		}
	}
}
