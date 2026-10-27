package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/report"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func runReportProcess(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], append([]string{"-test.run=^TestReportCommandProcess$", "--"}, args...)...)
	command.Dir = root
	command.Env = append(os.Environ(), reportCommandHelperEnv+"=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	return stdout.String(), stderr.String(), err
}

func TestReportExports(t *testing.T) {
	for _, name := range []string{"default", "txt", "md", "csv"} {
		t.Run(name, func(t *testing.T) {
			format := report.Format(name)
			if name == "default" {
				format = report.TXT
			}
			output := t.TempDir()
			// Both are powers of two on every vector width. Size 9 is only
			// registered as a boundary on some machines (for example width 8).
			args := []string{"-output", output, "-bench", "^Benchmark(Addition|Computation)$/^n=(8|16)$/^(Scalar|SIMD.*)$", "-count=2", "-benchtime=1x"}
			if name != "default" {
				args = append(args, "-format", name)
			}
			stdout, stderr, err := runReportProcess(t, args...)
			if err != nil {
				t.Fatalf("export failed: %v\n%s", err, stderr)
			}
			for _, group := range []string{"addition", "computation"} {
				files, err := os.ReadDir(filepath.Join(output, group))
				if err != nil {
					t.Fatal(err)
				}
				if len(files) != 2 {
					t.Fatalf("expected selected report and raw companion, got %v", files)
				}
				var rawPath, savedPath string
				for _, file := range files {
					path := filepath.Join(output, group, file.Name())
					switch filepath.Ext(path) {
					case ".json":
						rawPath = path
					case "." + string(format):
						savedPath = path
					default:
						t.Fatalf("unexpected output format: %s", path)
					}
				}
				if rawPath == "" || savedPath == "" || strings.TrimSuffix(rawPath, ".measurements.json") != strings.TrimSuffix(savedPath, "."+string(format)) {
					t.Fatalf("missing or unmatched artifact pair: %v", files)
				}
				raw, err := os.ReadFile(rawPath)
				if err != nil {
					t.Fatal(err)
				}
				var capture evidence.Capture
				if err := json.Unmarshal(raw, &capture); err != nil {
					t.Fatal(err)
				}
				run := capture.Run
				if capture.Schema != evidence.Schema || capture.Status != "complete" || run.BackendID != engine.BackendID || run.HarnessID != engine.HarnessID || capture.Machine.Lanes < 1 {
					t.Fatal("missing direct-result provenance")
				}
				wantCount := 16
				if group == "computation" {
					wantCount = 8
				}
				if len(run.Observations) != wantCount || !bytes.Contains(raw, []byte("-count=2")) || !bytes.Contains(raw, []byte("-benchtime=1x")) {
					t.Fatalf("raw measurements/configuration lost: %s", raw)
				}
				for _, observation := range run.Observations {
					if observation.Iterations != 1 || observation.DurationNS == nil || observation.NanosecondsPerOp != float64(*observation.DurationNS) || observation.Key.Case.Input.SHA256 == "" {
						t.Fatal("direct duration/iterations/input identity lost")
					}
				}
				summaries, err := result.Summarize(run.Observations)
				if err != nil {
					t.Fatal(err)
				}
				table, err := result.Compare(group, summaries)
				if err != nil {
					t.Fatal(err)
				}
				var selected, terminal bytes.Buffer
				if err := report.Write(&selected, format, table, ""); err != nil {
					t.Fatal(err)
				}
				if err := report.Write(&terminal, report.TXT, table, ""); err != nil {
					t.Fatal(err)
				}
				saved, err := os.ReadFile(savedPath)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.HasPrefix(saved, selected.Bytes()) || !strings.Contains(stdout, terminal.String()) {
					t.Fatalf("saved/terminal table differs from its own raw evidence\nsaved:\n%s\nstdout:\n%s", saved, stdout)
				}
				if format == report.CSV {
					if !bytes.Equal(saved, selected.Bytes()) {
						t.Fatal("CSV contains nontabular metadata")
					}
				} else if !bytes.Contains(saved, []byte("Measurement environment")) || !bytes.Contains(saved, []byte("build information:")) {
					t.Fatal("saved metadata missing")
				}
			}
			if strings.Contains(stdout, "| ELEMENTS |") || strings.Contains(stdout, "ELEMENTS,") {
				t.Fatal("saved format leaked into terminal presentation")
			}
		})
	}
}

func TestFailedReportPreservesRawEvidence(t *testing.T) {
	for _, tt := range []struct{ name, bench, benchtime, message string }{
		{"empty selection", "^NoSuchBenchmark$", "1x", "no benchmark measurements found"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output := t.TempDir()
			stdout, stderr, err := runReportProcess(t, "-group=addition", "-output", output, "-format=csv", "-bench", tt.bench, "-benchtime", tt.benchtime, "-count=1")
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("exit = %v, stderr = %s", err, stderr)
			}
			if !strings.Contains(stderr, tt.message) || !strings.Contains(stderr, "measurements saved to") || strings.Contains(stdout, "benchmark comparison") {
				t.Fatalf("failed measurement presented as success: stdout=%s stderr=%s", stdout, stderr)
			}
			files, err := os.ReadDir(filepath.Join(output, "addition"))
			if err != nil {
				t.Fatal(err)
			}
			if len(files) != 1 || !strings.HasSuffix(files[0].Name(), ".measurements.json") {
				t.Fatalf("failed measurement produced a report or lost raw output: %v", files)
			}
			data, err := os.ReadFile(filepath.Join(output, "addition", files[0].Name()))
			if err != nil {
				t.Fatal(err)
			}
			var capture evidence.Capture
			if err := json.Unmarshal(data, &capture); err != nil || capture.Status != "failed" || capture.Error == "" {
				t.Fatalf("invalid failed evidence: %s", data)
			}
		})
	}
}
