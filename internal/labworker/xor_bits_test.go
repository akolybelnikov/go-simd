package labworker

import (
	"context"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"math"
	"testing"
)

func TestRunnerOpaqueXorBits(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "xorbits", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "xor-bits-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	a, b, out := input.MakeXorBits(8)
	for i := range out {
		out[i] = math.Float32frombits(math.Float32bits(a[i]) ^ math.Float32bits(b[i]))
	}
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.Input.SHA256 != input.FingerprintFloat32(a, b) || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) {
			t.Fatal("worker XOR bits lost", o)
		}
	}
}
