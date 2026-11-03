package engine

import (
	"encoding/json"
	"math"
	"simd"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestTypedScalarRetentionAndNonfiniteEvidence(t *testing.T) {
	var value float32
	want := uint32(0x7fc12345)
	p := Prepared{Reduce: func() float32 { return math.Float32frombits(want) }, Scalar: &value, Bytes: 4}
	sample, err := (Testing{}).Measure(p, "3x")
	if err != nil || sample.Iterations != 3 || math.Float32bits(value) != want {
		t.Fatalf("scalar retention: %+v %v bits=%08x", sample, err, math.Float32bits(value))
	}
	if _, err := (Testing{}).Measure(Prepared{Reduce: p.Reduce}, "1x"); err == nil {
		t.Fatal("missing scalar destination accepted")
	}
	bits := math.Float32bits(value)
	capture := evidence.Capture{Schema: evidence.Schema, Status: "complete", Run: result.Run{Observations: []result.Observation{{ScalarResultBits: &bits}}}}
	var encoded strings.Builder
	if err := evidence.WriteCapture(&encoded, capture); err != nil {
		t.Fatal(err)
	}
	var decoded evidence.Capture
	if err := json.Unmarshal([]byte(encoded.String()), &decoded); err != nil || *decoded.Run.Observations[0].ScalarResultBits != want {
		t.Fatal("nonfinite scalar bits did not survive JSON")
	}
}

func TestReductionExecuteRetainsActualResultsAndFold(t *testing.T) {
	var vector simd.Float32s
	options := input.DefaultOptions()
	options.SumInput = "cancellation"
	plan, err := PrepareWithOptions("sum", Config{2, "3x", "^BenchmarkSum$/^n=64$"}, vector.Len(), options)
	if err != nil {
		t.Fatal(err)
	}
	run, err := Execute(plan, result.Run{ID: "sum", Environment: result.Environment{ID: "env"}, BackendID: BackendID, HarnessID: ReductionHarnessID}, Testing{})
	if err != nil || len(run.Observations) != 8 {
		t.Fatalf("%+v %v", run, err)
	}
	for _, o := range run.Observations {
		if o.ScalarResultBits == nil || o.Iterations != 3 || o.DurationNS == nil || o.Key.Case.Reduction.NativeLanes.Value != vector.Len() || o.Key.Case.Input.GeneratorID != "generated-sum-float32-v1/cancellation" {
			t.Fatalf("reduction identity lost: %+v", o)
		}
		want := uint32(0)
		if o.Label != "Scalar" {
			want = math.Float32bits(1)
		}
		if *o.ScalarResultBits != want {
			t.Fatalf("wrong returned sum for %s: %08x", o.Label, *o.ScalarResultBits)
		}
	}
}
