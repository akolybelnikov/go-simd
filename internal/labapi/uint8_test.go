package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"slices"
	"testing"
)

func TestSaturatingAdd8CatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("saturatingadd8", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Uint8 || op.Inputs[0].ElementBytes != 1 || op.Output.TypeID != catalog.Uint8 || op.LogicalBytesAtOne != 3 || !slices.Contains(op.SupportedSizes, 31) || !slices.Contains(op.SupportedSizes, 33) {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "saturatingadd8", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{31, 32, 33}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.TypeID != catalog.Uint8 || plan.Cases[0].LogicalBytes != 93 || plan.Input.GeneratorID != "generated-saturating-add-uint8-v1" {
		t.Fatal(plan, e)
	}
}
