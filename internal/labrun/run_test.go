package labrun

import (
	"context"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

func TestCompileRetainsOnlyConfirmedCases(t *testing.T) {
	machine := platform.Current()
	runtime := Runtime(machine)
	plan, err := labapi.ResolveExperimentPlan(labapi.PlanRequest{
		Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 1, Benchtime: "1x",
	}, runtime, labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := Compile(plan, machine, labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled.Cases) != 1 || compiled.Cases[0].Implementation.Descriptor.ID != catalog.Scalar || compiled.Cases[0].Elements != 8 {
		t.Fatalf("unexpected compiled cases: %+v", compiled.Cases)
	}
}

func TestExecuteCheckpointsAndReturnsCapture(t *testing.T) {
	machine := platform.Current()
	plan, err := labapi.ResolveExperimentPlan(labapi.PlanRequest{
		Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 2, Benchtime: "1x",
	}, Runtime(machine), labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	var checkpoints []int
	capture, err := Execute(context.Background(), "job-test", plan, labapi.DefaultPlanPolicy(), func(capture evidence.Capture) error {
		checkpoints = append(checkpoints, len(capture.Run.Observations))
		return nil
	})
	if err != nil || capture.Status != "complete" || len(capture.Run.Observations) != 2 || len(checkpoints) != 2 {
		t.Fatalf("status=%s observations=%d checkpoints=%v err=%v", capture.Status, len(capture.Run.Observations), checkpoints, err)
	}
}
