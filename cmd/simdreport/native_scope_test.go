package main

import (
	"bytes"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeGroupIsolationWithinSharedPackage(t *testing.T) {
	const selection = "^Benchmark.*$/^n=8$/^Scalar$"
	for _, group := range []string{"addition", "addition64", "bitroundtrip", "xorbits", "sum8", "clmul64", "copy64"} {
		t.Run(group, func(t *testing.T) {
			output := t.TempDir()
			stdout, stderr, e := runReportProcess(t, "-group="+group, "-backend=go-test", "-bench="+selection, "-count=1", "-benchtime=1x", "-output", output)
			if e != nil {
				t.Fatalf("%v\n%s\n%s", e, stdout, stderr)
			}
			files, _ := filepath.Glob(filepath.Join(output, group, "*.raw"))
			if len(files) != 1 {
				t.Fatal(files)
			}
			data, e := os.ReadFile(files[0])
			if e != nil {
				t.Fatal(e)
			}
			if !strings.Contains(string(data), "requested selection: "+selection) {
				t.Fatal("original selector lost")
			}
			run, e := result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "isolated", GeneratedInputs: true})
			if e != nil || len(run.Observations) != 1 {
				t.Fatal(run, e)
			}
			op, _ := catalog.ByGroup(group)
			if run.Observations[0].Key.Case.OperationID != op.ID {
				t.Fatal("sibling operation measured", run)
			}
		})
	}
}
