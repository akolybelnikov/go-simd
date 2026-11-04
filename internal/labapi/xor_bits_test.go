package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestXorBitsCatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("xorbits", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Float32 || len(op.Inputs) != 2 || op.LogicalBytesAtOne != 12 {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "xorbits", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.GeneratorID != "generated-xor-float32-bits-v1" || plan.Cases[0].LogicalBytes != 96 {
		t.Fatal(plan, e)
	}
}
