// Package labrun maps immutable laboratory plans to the callback-bearing
// benchmark engine. It is the only execution adapter used by isolated workers.
package labrun

import (
	"context"
	"fmt"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func Runtime(machine platform.Info) labapi.RuntimeContext {
	return labapi.RuntimeContext{
		GoVersion: machine.GoVersion, OS: machine.OS, Arch: machine.Arch,
		VectorBits: machine.Bits, Float32Lanes: machine.Lanes,
		GOMAXPROCS: machine.Procs, Emulated: machine.Emulated,
	}
}

func options(o labapi.ResolvedInputOptions) input.Options {
	return input.Options{
		FMAInput: o.FMAInput, DotInput: o.DotInput, SumInput: o.SumInput,
		Pattern: o.Pattern, Seed: o.Seed,
		Alpha: o.AlphaBits, Beta: o.BetaBits, Threshold: o.ThresholdBits,
		Lower: o.LowerBits, Upper: o.UpperBits,
	}
}

// Compile revalidates the plan against the current machine and policy before
// resolving executable callbacks. Only the exact confirmed cases are retained.
func Compile(plan labapi.ExperimentPlan, machine platform.Info, policy labapi.PlanPolicy) (engine.Plan, error) {
	if err := labapi.RevalidateExperimentPlan(plan, Runtime(machine), policy); err != nil {
		return engine.Plan{}, err
	}
	all, err := engine.PrepareWithOptions(plan.Group, engine.Config{Count: plan.Count, Benchtime: plan.Benchtime, Bench: "."}, machine.Lanes, options(plan.Input.Options))
	if err != nil {
		return engine.Plan{}, fmt.Errorf("compile confirmed plan: %w", err)
	}
	available := make(map[string]engine.Case, len(all.Cases))
	for _, c := range all.Cases {
		available[c.Name] = c
	}
	compiled := engine.Plan{Input: all.Input, Config: all.Config, Cases: make([]engine.Case, 0, len(plan.Cases))}
	seen := make(map[string]bool, len(plan.Cases))
	for _, requested := range plan.Cases {
		if seen[requested.Name] {
			return engine.Plan{}, fmt.Errorf("duplicate confirmed case %q", requested.Name)
		}
		seen[requested.Name] = true
		c, ok := available[requested.Name]
		if !ok || c.Operation.ID != requested.OperationID || c.Operation.TypeID != requested.TypeID ||
			c.Operation.ArithmeticID != requested.ArithmeticID || c.Implementation.Descriptor.ID != requested.StrategyID ||
			c.Elements != requested.Elements || c.Bytes != requested.LogicalBytes {
			return engine.Plan{}, fmt.Errorf("confirmed case %q is unavailable or inconsistent", requested.Name)
		}
		compiled.Cases = append(compiled.Cases, c)
	}
	return compiled, nil
}

// Execute returns a capture even on execution failure. observe receives a
// complete, valid partial capture after every successful observation.
func Execute(ctx context.Context, jobID string, plan labapi.ExperimentPlan, policy labapi.PlanPolicy, observe func(evidence.Capture) error) (evidence.Capture, error) {
	machine := platform.Current()
	compiled, err := Compile(plan, machine, policy)
	if err != nil {
		return evidence.Capture{}, err
	}
	op, ok := catalog.ByGroup(plan.Group)
	if !ok {
		return evidence.Capture{}, fmt.Errorf("unknown confirmed group %q", plan.Group)
	}
	if jobID == "" {
		return evidence.Capture{}, fmt.Errorf("job ID is required")
	}
	run := result.Run{
		ID:          jobID,
		Environment: result.Environment{ID: jobID + "/environment", GOOS: machine.OS, GOARCH: machine.Arch, CPU: machine.CPU},
		BackendID:   engine.BackendID, HarnessID: engine.HarnessFor(op), SourceFormat: evidence.Schema,
		Metadata: fmt.Sprintf("date: %s\nplan: %s\nmeasuring platform: %s/%s\n", time.Now().UTC().Format(time.RFC3339), plan.ID, machine.OS, machine.Arch),
	}
	base := evidence.Capture{
		Schema: evidence.Schema, Status: "complete", Group: plan.Group,
		Config:  evidence.Config{Input: &compiled.Input, Count: compiled.Config.Count, Benchtime: compiled.Config.Benchtime, Bench: compiled.Config.Bench},
		Machine: machine,
	}
	checkpoint := func(snapshot result.Run) error {
		if observe == nil {
			return nil
		}
		partial := base
		partial.Status = "failed"
		partial.Error = "measurement interrupted before completion"
		partial.Run = snapshot
		return observe(partial)
	}
	run, runErr := engine.ExecuteObserved(ctx, compiled, run, engine.Testing{}, checkpoint)
	base.Run = run
	if runErr != nil {
		base.Status = "failed"
		base.Error = runErr.Error()
	}
	return base, runErr
}
