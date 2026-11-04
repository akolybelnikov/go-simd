package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"testing"
)

func TestRunnerClmul64(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "clmul64", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "clmul64-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	a, b, out := input.MakeClmul64(8)
	if e := integer.Clmul64Scalar(a, b, out); e != nil {
		t.Fatal(e)
	}
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.Input.SHA256 != input.FingerprintUint64(a, b) || o.ArrayResultSHA256 != input.FingerprintUint64(out, nil) {
			t.Fatal("worker XOR bits lost", o)
		}
	}
}
