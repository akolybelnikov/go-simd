package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestSum8CatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("sum8", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Float32 || len(op.Inputs) != 1 || op.LogicalBytesAtOne != 8 {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "sum8", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.GeneratorID != "generated-sum8-float32-v1/balanced" || plan.Cases[0].LogicalBytes != 36 || plan.Cases[0].OutputElements != 1 {
		t.Fatal(plan, e)
	}
}
