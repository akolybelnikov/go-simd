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
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"
)

func TestBitRoundTripResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("bitroundtrip")
	sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 1024: true}
	selection := fmt.Sprintf("^BenchmarkBitRoundTrip$/^n=(1|%d|%d|%d|1024)$", w-1, w, w+1)
	var runs []result.Run
	for _, backend := range []string{"compiled", "go-test"} {
		output := t.TempDir()
		stdout, stderr, e := runReportProcess(t, "-group=bitroundtrip", "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output)
		if e != nil {
			t.Fatalf("%s %v\n%s\n%s", backend, e, stdout, stderr)
		}
		files, _ := filepath.Glob(filepath.Join(output, "bitroundtrip", "*"))
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
				if c.Status != "complete" || run.HarnessID != engine.BitRoundTripHarnessID {
					t.Fatal(c)
				}
			} else if strings.HasSuffix(file, ".raw") {
				run, e = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
				if e != nil || run.HarnessID != result.LegacyBitRoundTripHarnessID {
					t.Fatal(run, e)
				}
			}
		}
		if len(run.Observations) != len(sizes)*4 {
			t.Fatal(len(run.Observations))
		}
		for _, o := range run.Observations {
			c := o.Key.Case
			want := input.FingerprintForOperation(op, c.Elements, input.DefaultOptions())
			if !sizes[c.Elements] || c.OperationID != catalog.BitRoundTrip || c.TypeID != catalog.Float32 || c.Parameters != "" || c.Input.GeneratorID != input.DefaultOptions().Generator(op) || c.Input.SHA256 != want || o.ArrayResultSHA256 != want || c.Vector.NativeLanes.Value != w || c.Copy != (result.CopyIdentity{}) || o.Iterations != 3 {
				t.Fatal("roundtrip evidence lost", o)
			}
		}
		runs = append(runs, run)
	}
	for i, a := range runs[0].Observations {
		b := runs[1].Observations[i]
		if a.Key.Case != b.Key.Case || a.ArrayResultSHA256 != b.ArrayResultSHA256 {
			t.Fatal("roundtrip backend drift", a, b)
		}
	}
}
