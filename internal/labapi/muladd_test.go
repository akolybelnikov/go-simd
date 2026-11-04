package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestMulAddCatalogueAndPlanContract(t *testing.T) {
	response, e := DescribeCatalog("muladd", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if len(op.Inputs) != 3 || op.Inputs[2].Name != "c" || op.LogicalBytesAtOne != 16 || len(op.Parameters) != 0 {
		t.Fatal(op)
	}
	fixture := "rounding"
	alpha := uint32(123)
	req := PlanRequest{Group: "muladd", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 2, Benchtime: "3x", Input: &InputOptions{FMAInput: &fixture, AlphaBits: &alpha}}
	plan, e := ResolveExperimentPlan(req, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.Options.FMAInput != "rounding" || plan.Input.GeneratorID != "generated-muladd-float32-v1/rounding" || len(plan.Input.Parameters) != 0 || plan.Cases[0].LogicalBytes != 128 {
		t.Fatal(plan, e)
	}
	req.Input.AlphaBits = nil
	other, e := ResolveExperimentPlan(req, testRuntime(), DefaultPlanPolicy())
	if e != nil || other.ID != plan.ID {
		t.Fatal("irrelevant parameter changes plan")
	}
	fixture = "balanced"
	different, e := ResolveExperimentPlan(req, testRuntime(), DefaultPlanPolicy())
	if e != nil || different.ID == plan.ID {
		t.Fatal("fixture missing from plan")
	}
}
