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

func TestRunnerCopy64(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "copy64", Variants: []string{catalog.Scalar, catalog.GoCopy, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "copy64-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 3 {
		t.Fatal(got, e)
	}
	x, _ := input.MakeCopy64(8)
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.Input.SHA256 != input.FingerprintFloat64(x, nil) || o.ArrayResultSHA256 != input.FingerprintFloat64(x, nil) {
			t.Fatal("worker copy64 lost", o)
		}
	}
}
