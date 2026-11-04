package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestCopy64CatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("copy64", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Float64 || len(op.Inputs) != 1 || op.LogicalBytesAtOne != 16 {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "copy64", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.GeneratorID != "generated-copy-float64-bits-v1" || plan.Cases[0].LogicalBytes != 128 {
		t.Fatal(plan, e)
	}
}
