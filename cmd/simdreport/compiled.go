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
