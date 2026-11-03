package main

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestPackedPredicateResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("predicatepack")
	sizes := map[int]bool{1: true, 7: true, 8: true, 9: true, 63: true, 64: true, 65: true, w - 1: true, w: true, w + 1: true}
	selection := fmt.Sprintf("^BenchmarkPredicatePacked$/^n=(1|7|8|9|63|64|65|%d|%d|%d)$", w-1, w, w+1)
	for _, pattern := range []string{"none", "all", "alternating", "clustered", "shuffled"} {
		var runs []result.Run
		o := input.DefaultOptions()
		o.Pattern = pattern
		o.Seed = 123
		o.Threshold = math.Float32bits(-1.5)
		for _, backend := range []string{"compiled", "go-test"} {
			output := t.TempDir()
			stdout, stderr, err := runReportProcess(t, "-group=predicatepack", "-pattern="+pattern, "-seed=123", "-threshold=-1.5", "-backend="+backend, "-bench="+selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output)
			if err != nil {
				t.Fatalf("%s/%s: %v\n%s\n%s", pattern, backend, err, stdout, stderr)
			}
			files, _ := filepath.Glob(filepath.Join(output, "predicatepack", "*"))
			if len(files) != 2 {
				t.Fatal("wrong artifacts", files)
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
					if run.HarnessID != engine.PackedPredicateHarnessID || c.Status != "complete" {
						t.Fatal(c)
					}
				} else if strings.HasSuffix(file, ".raw") {
					run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
					if err != nil || run.HarnessID != result.LegacyPackedPredicateHarnessID {
						t.Fatal(err, run)
					}
				}
			}
			if len(run.Observations) != len(sizes)*4 {
				t.Fatal("wrong observation count", len(run.Observations))
			}
			for _, obs := range run.Observations {
				c := obs.Key.Case
				x, _, _ := input.MakeForOperation(op, c.Elements, o)
				expected := make([]byte, (len(x)+7)/8)
				for i, v := range x {
					if v > o.Float("threshold") {
						expected[i/8] |= 1 << uint(i%8)
					}
				}
				digest := sha256.Sum256(expected)
				if !sizes[c.Elements] || obs.Iterations != 3 || c.OperationID != catalog.PredicateGTPacked || c.ArithmeticID != op.ArithmeticID || c.Parameters != o.Identity(op) || c.Input.SHA256 != input.FingerprintFloat32(x, nil) || c.Predicate.EncodingID != result.PackedPredicateEncodingID || c.Predicate.NativeLanes.Value != w || obs.ByteResultSHA256 != fmt.Sprintf("%x", digest) || obs.ScalarResultBits != nil || obs.ArrayResultLastBits != nil {
					t.Fatal("mask output/context lost", obs)
				}
			}
			runs = append(runs, run)
		}
		for i, a := range runs[0].Observations {
			b := runs[1].Observations[i]
			if a.Key.Case != b.Key.Case || a.ByteResultSHA256 != b.ByteResultSHA256 {
				t.Fatal("mask backend parity")
			}
		}
	}
	output := t.TempDir()
	_, _, err := runReportProcess(t, "-group=predicatepack", "-threshold=+Inf", "-pattern=all", "-output", output)
	files, _ := os.ReadDir(output)
	if err == nil || len(files) != 0 {
		t.Fatal("impossible fixture measured")
	}
}
