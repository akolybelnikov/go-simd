package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"testing"
)

func TestRunnerBitRoundTripOpaqueBits(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "bitroundtrip", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "opaque-bits-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	op, _ := catalog.ByGroup("bitroundtrip")
	want := input.FingerprintForOperation(op, 8, input.DefaultOptions())
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.Input.SHA256 != want || o.ArrayResultSHA256 != want {
			t.Fatal("worker changed payload bits", o)
		}
	}
}
