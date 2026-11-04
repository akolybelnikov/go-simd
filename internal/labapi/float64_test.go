package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"slices"
	"testing"
)

func TestFloat64CataloguePlanAndWidth(t *testing.T) {
	response, e := DescribeCatalog("addition64", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Float64 || op.Inputs[0].ElementBytes != 8 || op.LogicalBytesAtOne != 24 || !slices.Contains(op.SupportedSizes, 3) || !slices.Contains(op.SupportedSizes, 5) {
		t.Fatal("actual float64 boundary", op)
	}
	req := PlanRequest{Group: "addition64", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{3, 4, 5}, Count: 1, Benchtime: "1x"}
	plan, e := ResolveExperimentPlan(req, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.TypeID != catalog.Float64 || plan.Input.GeneratorID != "generated-addition-float64-v1" || len(plan.Cases) != 6 || plan.Cases[0].LogicalBytes != 72 {
		t.Fatal(plan, e)
	}
}
