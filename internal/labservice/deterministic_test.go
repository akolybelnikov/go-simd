package labservice

import (
	"bytes"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

func TestDeterministicCatalogAndPlanAgree(t *testing.T) {
	machine := platform.Current()
	service, err := NewDeterministic(machine, labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	catalogue, err := service.Catalog("addition")
	if err != nil {
		t.Fatal(err)
	}
	op := catalogue.Operations[0]
	plan, err := service.Plan(labapi.PlanRequest{Group: op.Group, Variants: []string{catalog.Scalar}, Sizes: []int{op.SupportedSizes[0]}, Count: 3, Benchtime: "1x"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Runtime != labrun.Runtime(machine) || plan.Cases[0].Elements != op.SupportedSizes[0] {
		t.Fatalf("catalogue/plan mismatch: %+v %+v", op, plan)
	}
}

func TestDeterministicInspectUsesBoundedReader(t *testing.T) {
	service, err := NewDeterministic(platform.Current(), labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	limits := evidence.DefaultReadLimits()
	limits.MaxBytes = 4
	if _, err := service.Inspect(bytes.NewBufferString("not json"), limits); err == nil {
		t.Fatal("invalid evidence accepted")
	}
}
