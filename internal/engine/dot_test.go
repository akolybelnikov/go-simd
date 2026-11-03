package engine

import (
	"errors"
	"math"
	"simd"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestCheckedScalarAdapterRetainsValueAndPropagatesError(t *testing.T) {
	var value float32
	want := uint32(0xff800000)
	call := func() (float32, error) { return math.Float32frombits(want), nil }
	p := Prepared{ReduceChecked: call, Scalar: &value, Bytes: 12}
	sample, err := (Testing{}).Measure(p, "3x")
	if err != nil || sample.Iterations != 3 || math.Float32bits(value) != want {
		t.Fatal("checked scalar result lost", sample, err)
	}
	value = 123
	p.ReduceChecked = func() (float32, error) { return reduction.DotScalar([]float32{1}, nil) }
	if _, err := (Testing{}).Measure(p, "2x"); !errors.Is(err, reduction.ErrLengthMismatch) || value != 123 {
		t.Fatal("checked error/retention contract lost", err)
	}
	p.Reduce = func() float32 { return 0 }
	if _, err := (Testing{}).Measure(p, "1x"); err == nil {
		t.Fatal("ambiguous scalar callbacks accepted")
	}
	if _, err := (Testing{}).Measure(Prepared{ReduceChecked: call}, "1x"); err == nil {
		t.Fatal("missing scalar destination accepted")
	}
}

type dotFailingBackend struct{ calls int }

func (b *dotFailingBackend) Measure(p Prepared, _ string) (Sample, error) {
	b.calls++
	if b.calls == 2 {
		return Sample{}, errors.New("injected dot failure")
	}
	value, err := p.ReduceChecked()
	if err != nil {
		return Sample{}, err
	}
	*p.Scalar = value
	return Sample{Iterations: 3, Duration: time.Nanosecond}, nil
}

func TestDotExecuteEvidenceAndPartialFailure(t *testing.T) {
	var vector simd.Float32s
	o := input.DefaultOptions()
	o.DotInput = "cancellation"
	plan, err := PrepareWithOptions("dot", Config{2, "3x", "^BenchmarkDot$/^n=64$"}, vector.Len(), o)
	if err != nil {
		t.Fatal(err)
	}
	run := result.Run{ID: "dot", Environment: result.Environment{ID: "env"}, BackendID: BackendID, HarnessID: DotHarnessID}
	complete, err := Execute(plan, run, Testing{})
	if err != nil || len(complete.Observations) != 8 {
		t.Fatal("dot execution failed", err)
	}
	for _, observation := range complete.Observations {
		c := observation.Key.Case
		want := uint32(0)
		if observation.Label != "Scalar" {
			want = math.Float32bits(1)
		}
		if observation.ScalarResultBits == nil || *observation.ScalarResultBits != want || c.Input.SHA256 != input.FingerprintForOperation(plan.Cases[0].Operation, 64, o) || c.Input.GeneratorID != "generated-dot-float32-v1/cancellation" || c.Reduction.NativeLanes.Value != vector.Len() || observation.Iterations != 3 || observation.DurationNS == nil {
			t.Fatal("dot evidence incomplete", observation)
		}
	}
	partial, err := Execute(plan, run, &dotFailingBackend{})
	if err == nil || len(partial.Observations) != 1 || partial.Observations[0].ScalarResultBits == nil {
		t.Fatal("dot partial observations lost", err)
	}
}
