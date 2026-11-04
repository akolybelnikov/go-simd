package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestClmul64CatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("clmul64", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Uint64 || len(op.Inputs) != 2 || op.LogicalBytesAtOne != 24 {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "clmul64", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.GeneratorID != "generated-clmul-uint64-v1" || plan.Cases[0].LogicalBytes != 192 {
		t.Fatal(plan, e)
	}
}
