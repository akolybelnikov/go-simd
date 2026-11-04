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

func TestSaturatingAdd8ResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Uint8s
	w := vector.Len()
	op, _ := catalog.ByGroup("saturatingadd8")
	sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 1024: true}
	selection := fmt.Sprintf("^BenchmarkSaturatingAdd8$/^n=(1|%d|%d|%d|1024)$", w-1, w, w+1)
	var runs []result.Run
	for _, backend := range []string{"compiled", "go-test"} {
		output := t.TempDir()
		stdout, stderr, e := runReportProcess(t, "-group=saturatingadd8", "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output)
		if e != nil {
			t.Fatalf("%s %v\n%s\n%s", backend, e, stdout, stderr)
		}
		files, _ := filepath.Glob(filepath.Join(output, "saturatingadd8", "*"))
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
				if c.Status != "complete" || run.HarnessID != engine.SaturatingAdd8HarnessID {
					t.Fatal(c)
				}
			} else if strings.HasSuffix(file, ".raw") {
				run, e = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
				if e != nil || run.HarnessID != result.LegacySaturatingAdd8HarnessID {
					t.Fatal(run, e)
				}
			}
		}
		if len(run.Observations) != len(sizes)*4 {
			t.Fatal(len(run.Observations))
		}
		for _, o := range run.Observations {
			c := o.Key.Case
			a, b, out := input.MakeSaturatingAdd8(c.Elements)
			for i := range out {
				out[i] = uint8(min(int(a[i])+int(b[i]), 255))
			}
			if !sizes[c.Elements] || c.OperationID != catalog.SaturatingAdd8 || c.TypeID != catalog.Uint8 || c.ArithmeticID != op.ArithmeticID || c.Parameters != "" || c.Input.GeneratorID != input.SaturatingAdd8GeneratorID || c.Input.SHA256 != input.FingerprintUint8(a, b) || o.ByteResultSHA256 != input.FingerprintUint8(out, nil) || c.Vector.NativeLanes.Value != w || !c.Vector.NativeLanes.Known || c.Predicate != (result.PredicateIdentity{}) || o.ArrayResultSHA256 != "" || o.Iterations != 3 {
				t.Fatal("saturation evidence lost", o)
			}
		}
		runs = append(runs, run)
	}
	for i, a := range runs[0].Observations {
		b := runs[1].Observations[i]
		if a.Key.Case != b.Key.Case || a.ByteResultSHA256 != b.ByteResultSHA256 {
			t.Fatal("saturation backend drift", a, b)
		}
	}
}
