package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"testing"
)

func TestRunnerThreeArrayMulAdd(t *testing.T) {
	fixture := "rounding"
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "muladd", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x", Input: &labapi.InputOptions{FMAInput: &fixture}}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "ternary-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	options := input.DefaultOptions()
	options.FMAInput = fixture
	a, b, c, out := input.MakeMulAdd(8, options)
	for i := range out {
		out[i] = fma32.FMA(a[i], b[i], c[i])
	}
	for _, o := range got.Capture.Run.Observations {
		if o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || o.Key.Case.Input.SHA256 != input.FingerprintTernaryFloat32(a, b, c) {
			t.Fatal("worker third operand lost", o)
		}
	}
}
