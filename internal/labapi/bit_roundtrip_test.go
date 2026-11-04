package labapi

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestBitRoundTripCatalogueAndPlan(t *testing.T) {
	response, e := DescribeCatalog("bitroundtrip", testRuntime(), "software-fma32/v1")
	if e != nil || len(response.Operations) != 1 {
		t.Fatal(response, e)
	}
	op := response.Operations[0]
	if op.TypeID != catalog.Float32 || len(op.Inputs) != 1 || op.LogicalBytesAtOne != 8 {
		t.Fatal(op)
	}
	plan, e := ResolveExperimentPlan(PlanRequest{Group: "bitroundtrip", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, testRuntime(), DefaultPlanPolicy())
	if e != nil || plan.Input.GeneratorID != "generated-bit-roundtrip-float32-v1" || plan.Cases[0].LogicalBytes != 64 {
		t.Fatal(plan, e)
	}
}
