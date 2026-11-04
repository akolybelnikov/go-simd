package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"simd"
	"testing"
)

func TestRunnerTypedFloat64Addition(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "addition64", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "binary64-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	a, b, out := input.MakeAddition64(8)
	for i := range out {
		out[i] = a[i] + b[i]
	}
	var vector simd.Float64s
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.TypeID != catalog.Float64 || o.ArrayResultSHA256 != input.FingerprintFloat64(out, nil) || o.Key.Case.Input.SHA256 != input.FingerprintFloat64(a, b) || o.Key.Case.Vector.NativeLanes.Value != vector.Len() {
			t.Fatal("worker binary64 lost", o)
		}
	}
}
