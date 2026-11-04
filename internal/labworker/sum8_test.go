package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"testing"
)

func TestRunnerSum8(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "sum8", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{9}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "sum8-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	op, _ := catalog.ByGroup("sum8")
	a, b, out := input.MakeForOperation(op, 9, input.DefaultOptions())
	if e := reduction.Sum8SIMD(a, out); e != nil {
		t.Fatal(e)
	}
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.OutputElements.Value != 2 || o.Key.Case.Input.SHA256 != input.FingerprintFloat32(a, b) || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) {
			t.Fatal("worker sum8 lost", o)
		}
	}
}
