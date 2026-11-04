package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/report"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func measureCompiledReport(group string, plan engine.Plan, machine platform.Info, output string, format report.Format, stdout io.Writer) error {
	directory := filepath.Join(output, group)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	captureFile, err := os.CreateTemp(directory, "benchmark-"+time.Now().UTC().Format("20060102-150405")+"-*.measurements.json")
	if err != nil {
		return err
	}
	id := strings.TrimSuffix(filepath.Base(captureFile.Name()), ".measurements.json")
	op, _ := catalog.ByGroup(group)
	harness := engine.HarnessFor(op)
	build, _ := json.Marshal(machine.Build)
	metadata := fmt.Sprintf("date: %s\ngo: %s\nplatform: %s/%s\ncpu: %s\nGOMAXPROCS: %d\nSIMD bits: %d\nSIMD float32 lanes: %d\nemulated: %t\nsource revision: %s\nsource modified: %s\nbackend: %s\nharness: %s\ngroup: %s\nselection: %q\n-count=%d\n-benchtime=%s\nbuild information: %s\n",
		time.Now().UTC().Format(time.RFC3339), machine.GoVersion, machine.OS, machine.Arch,
		machine.CPU, machine.Procs, machine.Bits, machine.Lanes, machine.Emulated,
		machine.SourceRevision, machine.SourceModified, engine.BackendID, harness,
		group, plan.Config.Bench, plan.Config.Count, plan.Config.Benchtime, build)
	run := result.Run{
		ID: id, Environment: result.Environment{ID: id + "/environment", GOOS: machine.OS, GOARCH: machine.Arch, CPU: machine.CPU},
		BackendID: engine.BackendID, HarnessID: harness,
		SourceFormat: evidence.Schema, Metadata: metadata,
	}
	if op.Parameterized() {
		run.Metadata += "parameters: " + plan.Input.Identity(op) + "\nfixture: " + plan.Input.Generator(op) + "\n"
	}
	if op.ID == catalog.Copy {
		run.Metadata += "fixture: " + plan.Input.Generator(op) + "\ncopy buffers: disjoint; exact in-place API supported; partial overlap unsupported\noutput: exact float32 bits; full ArrayResultSHA256 retained after timing\nlogical traffic: 8*n bytes/call\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if op.ID == catalog.Clamp {
		run.Metadata += "output: ordered clamp preserves input ties/NaN bits and exact clipping-bound bits; full ArrayResultSHA256 retained after timing\nlogical traffic: 8*n bytes/call\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if op.ID == catalog.Add64 {
		run.Metadata += "fixture: " + plan.Input.Generator(op) + "\nfloat64 input/output; full binary64 output digest after timing; logical traffic 24*n\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if op.ID == catalog.MulAdd {
		run.Metadata += "fixture: " + plan.Input.Generator(op) + "\nthree streamed input arrays; full ArrayResultSHA256 after timing; logical traffic 16*n\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if op.ID == catalog.AffineFMA || op.ID == catalog.MulAdd {
		run.Metadata += "FMA32 SIMD path: " + machine.FMA32Path + "\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if op.ID == catalog.DotFMA {
		run.Metadata += "FMA32 vector gate: " + machine.FMA32Path + "; per-case path and fold retained\n"
	}
	if op.ScalarResult() {
		run.Metadata += "fixture: " + plan.Input.Generator(op) + "\nnumerics: " + op.NumericalPolicy + fmt.Sprintf("\nlogical traffic: %d*n + 4 bytes/call\n", 4*len(op.Inputs))
	}
	if op.ID == catalog.PredicateGTPacked {
		run.Metadata += "output: uint8-packed-lsb0/v1; unused high bits zero; full ByteResultSHA256 retained after timing\nlogical traffic: 4*n+ceil(n/8) bytes/call\nnumerics: " + op.NumericalPolicy + "\n"
	} else if op.ByteResult() {
		run.Metadata += "output: uint8-zero-one/v1; full ByteResultSHA256 retained after timing\nlogical traffic: 5*n bytes/call\nnumerics: " + op.NumericalPolicy + "\n"
	}
	if _, err := fmt.Fprintf(stdout, "Measuring %s (%d repetitions; compiled backend)...\n", group, plan.Config.Count); err != nil {
		captureFile.Close()
		return err
	}
	run, runErr := engine.Execute(plan, run, engine.Testing{})
	capture := evidence.Capture{
		Schema: evidence.Schema, Status: "complete", Group: group,
		Config:  evidence.Config{Count: plan.Config.Count, Benchtime: plan.Config.Benchtime, Bench: plan.Config.Bench, Input: &plan.Input},
		Machine: machine, Run: run,
	}
	if runErr != nil {
		capture.Status = "failed"
		capture.Error = runErr.Error()
	}
	writeErr := evidence.WriteCapture(captureFile, capture)
	closeErr := captureFile.Close()
	if err := errors.Join(runErr, writeErr, closeErr); err != nil {
		return fmt.Errorf("benchmark failed (measurements saved to %s): %w", captureFile.Name(), err)
	}
	return saveReport(group, captureFile.Name(), run, format, stdout)
}
