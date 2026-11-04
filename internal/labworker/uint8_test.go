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

func TestRunnerSaturatingAdd8TypedBytes(t *testing.T) {
	plan, e := labapi.ResolveExperimentPlan(labapi.PlanRequest{Group: "saturatingadd8", Variants: []string{catalog.Scalar, catalog.SIMDPartial}, Sizes: []int{8}, Count: 1, Benchtime: "1x"}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if e != nil {
		t.Fatal(e)
	}
	got, e := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), Request{Schema: RequestSchema, JobID: "uint8-job", Plan: plan})
	if e != nil || got.Capture == nil || got.Manifest.Status != "complete" || len(got.Capture.Run.Observations) != 2 {
		t.Fatal(got, e)
	}
	a, b, out := input.MakeSaturatingAdd8(8)
	for i := range out {
		out[i] = uint8(min(int(a[i])+int(b[i]), 255))
	}
	var vector simd.Uint8s
	for _, o := range got.Capture.Run.Observations {
		if o.Key.Case.TypeID != catalog.Uint8 || o.Key.Case.Input.SHA256 != input.FingerprintUint8(a, b) || o.ByteResultSHA256 != input.FingerprintUint8(out, nil) || o.Key.Case.Vector.NativeLanes.Value != vector.Len() {
			t.Fatal("worker byte inputs lost", o)
		}
	}
}
