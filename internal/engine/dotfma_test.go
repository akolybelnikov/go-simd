package engine

import (
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestFusedDotExecutionAndPartialCapture(t *testing.T) {
	var vector simd.Float32s
	o := input.DefaultOptions()
	o.DotInput = "fusion-sensitive"
	p, err := PrepareWithOptions("dotfma", Config{2, "3x", "^BenchmarkDotFMA$/^n=64$"}, vector.Len(), o)
	if err != nil {
		t.Fatal(err)
	}
	run := result.Run{ID: "fused", Environment: result.Environment{ID: "env"}, BackendID: BackendID, HarnessID: DotFMAHarnessID}
	complete, err := Execute(p, run, Testing{})
	if err != nil || len(complete.Observations) != 8 {
		t.Fatal("fused execute", err)
	}
	for _, obs := range complete.Observations {
		if obs.ScalarResultBits == nil || obs.Iterations != 3 || obs.DurationNS == nil || obs.Key.Case.FMA.RoundingID != result.FusedDotRoundingID || obs.Key.Case.FMA.NativeLanes != obs.Key.Case.Reduction.NativeLanes || obs.Key.Case.Input.GeneratorID != "generated-dot-fma-float32-v1/fusion-sensitive" {
			t.Fatal("fused execution facts lost", obs)
		}
	}
	partial, err := Execute(p, run, &dotFailingBackend{})
	if err == nil || len(partial.Observations) != 1 || partial.Observations[0].ScalarResultBits == nil {
		t.Fatal("partial fused capture lost")
	}
}
