package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestStandaloneBinaryWithoutGoOrSource(t *testing.T) {
	buildDir, workDir := t.TempDir(), t.TempDir()
	name := "simdreport"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(buildDir, name)
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	invoke := func(args ...string) (string, string, error) {
		command := exec.Command(binary, args...)
		command.Dir = workDir
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if !strings.EqualFold(key, "PATH") && !strings.EqualFold(key, "GOEXPERIMENT") {
				command.Env = append(command.Env, entry)
			}
		}
		command.Env = append(command.Env, "PATH=", "GOEXPERIMENT=")
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		err := command.Run()
		return stdout.String(), stderr.String(), err
	}
	if stdout, stderr, err := invoke("-list"); err != nil || !strings.Contains(stdout, "[compute-expression/float32]") {
		t.Fatalf("list: %v\n%s\n%s", err, stdout, stderr)
	}
	if files, _ := os.ReadDir(workDir); len(files) != 0 {
		t.Fatal("listing created artifacts")
	}
	if stdout, stderr, err := invoke("-bench=^Benchmark(Addition|Computation)$/^n=(8|16)$", "-count=2", "-benchtime=2x", "-format=md"); err != nil || !strings.Contains(stdout, "benchmark comparison") {
		t.Fatalf("report: %v\n%s\n%s", err, stdout, stderr)
	}
	for _, group := range []string{"addition", "computation"} {
		paths, err := filepath.Glob(filepath.Join(workDir, "artifacts", group, "*.measurements.json"))
		if err != nil || len(paths) != 1 {
			t.Fatal(paths, err)
		}
		data, err := os.ReadFile(paths[0])
		if err != nil {
			t.Fatal(err)
		}
		var capture evidence.Capture
		if err := json.Unmarshal(data, &capture); err != nil || capture.Status != "complete" || capture.Machine.OS != runtime.GOOS || capture.Machine.GoVersion == "" {
			t.Fatalf("bad capture: %v %s", err, data)
		}
		for _, observation := range capture.Run.Observations {
			if observation.Iterations != 2 || observation.DurationNS == nil {
				t.Fatal("fixed iteration evidence lost")
			}
		}
		if _, err := os.Stat(strings.TrimSuffix(paths[0], ".measurements.json") + ".md"); err != nil {
			t.Fatal(err)
		}
	}
	_, stderr, err := invoke("-backend=go-test", "-group=addition", "-count=1", "-benchtime=1x", "-bench=^BenchmarkAddition$/^n=1$/^Scalar$", "-output=developer")
	if err == nil || !strings.Contains(stderr, "go env") {
		t.Fatalf("developer backend unexpectedly fell back: %v %s", err, stderr)
	}
}

func TestSharedListingAndAliases(t *testing.T) {
	for _, args := range [][]string{{"-list"}, {"-list", "-function=add", "-group=addition", "-addition"}, {"-list", "-computation"}, {"-list", "-addition=false"}} {
		var output bytes.Buffer
		if err := run(args, &output); err != nil || !strings.Contains(output.String(), "no benchmarks run") {
			t.Fatalf("%v: %v %s", args, err, &output)
		}
		if slices.Contains(args, "-computation") && strings.Contains(output.String(), "[add/float32]") {
			t.Fatal("alias filtering drift")
		}
	}
	if err := run([]string{"-list", "-addition", "-computation"}, &bytes.Buffer{}); err == nil {
		t.Fatal("conflicting aliases accepted")
	}
}

func TestExplicitDeveloperBackend(t *testing.T) {
	output := t.TempDir()
	stdout, stderr, err := runReportProcess(t, "-backend=go-test", "-group=addition", "-bench=^BenchmarkAddition$/^n=8$/^Scalar$", "-count=2", "-benchtime=1x", "-output", output)
	if err != nil {
		t.Fatalf("%v\n%s\n%s", err, stdout, stderr)
	}
	paths, _ := filepath.Glob(filepath.Join(output, "addition", "*.raw"))
	if len(paths) != 1 {
		t.Fatal(paths)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	run, err := result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "developer"})
	if err != nil || len(run.Observations) != 2 || run.BackendID != result.LegacyBackendID || run.Observations[0].DurationNS != nil {
		t.Fatalf("legacy evidence changed: %+v %v", run, err)
	}
}

func TestSelectionMatchesNativeGo(t *testing.T) {
	for _, expression := range []string{
		"^Benchmark(Addition|Computation)$/^n=(8|16)$/^(Scalar|SIMD.*)$",
		"^BenchmarkAddition$/^n=8$/^Scalar$|^BenchmarkComputation$/^n=16$/^SIMD$",
		"^BenchmarkAddition$/^n=[8/9]$/^Scalar$",
		"^BenchmarkAddition$/^n=8$/^Scalar$/extra",
	} {
		t.Run(expression, func(t *testing.T) {
			command := exec.Command("go", "test", "./internal/addition", "./internal/computation", "-run=^$", "-bench="+expression, "-benchtime=1x")
			command.Dir = filepath.Join("..", "..")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("native comparison: %v\n%s", err, output)
			}
			var native, compiled []string
			for _, line := range strings.Split(string(output), "\n") {
				fields := strings.Fields(line)
				if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
					continue
				}
				name := fields[0]
				if pos := strings.LastIndex(name, "-"); pos >= 0 {
					name = name[:pos]
				}
				native = append(native, name)
			}
			for _, group := range []string{"addition", "computation"} {
				plan, err := engine.Prepare(group, engine.Config{Count: 1, Benchtime: "1x", Bench: expression}, platform.Current().Lanes)
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range plan.Cases {
					compiled = append(compiled, c.Name)
				}
			}
			slices.Sort(native)
			slices.Sort(compiled)
			if !slices.Equal(native, compiled) {
				t.Fatalf("selector drift:\nnative=%v\ncompiled=%v\n%s", native, compiled, output)
			}
		})
	}
}
