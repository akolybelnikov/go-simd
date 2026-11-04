package analysis

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestInterpretCaptureProducesTraceableFactsAndBoundaries(t *testing.T) {
	baseKey := result.MeasurementKey{
		RunID: "run", EnvironmentID: "env", BackendID: "backend", HarnessID: "harness",
		Case: result.Case{OperationID: catalog.Add, TypeID: "float32", StrategyID: catalog.Scalar, Elements: 8, Input: result.InputIdentity{GeneratorID: "fixture", SHA256: "same"}, ExecutionMode: "serial"},
	}
	baseKey.Case.Workers = result.OptionalInt{Value: 1, Known: true}
	vectorKey := baseKey
	vectorKey.Case.StrategyID = catalog.SIMDPartial
	vectorKey.Case.Workers = result.OptionalInt{Value: 4, Known: true}
	capture := evidence.Capture{
		Schema: evidence.Schema, Status: "complete", Group: "addition",
		Config:  evidence.Config{Count: 3, Benchtime: "1x", Bench: "."},
		Machine: evidence.Machine{GoVersion: "go-test", OS: "linux", Arch: "amd64", CPU: "test", Bits: 256, Lanes: 8, Procs: 4},
		Run: result.Run{ID: "run", Observations: []result.Observation{
			{Key: baseKey, Label: "Scalar", Repetition: 1, NanosecondsPerOp: 10, SourceBenchmark: "scalar"},
			{Key: baseKey, Label: "Scalar", Repetition: 2, NanosecondsPerOp: 12, SourceBenchmark: "scalar"},
			{Key: baseKey, Label: "Scalar", Repetition: 3, NanosecondsPerOp: 11, SourceBenchmark: "scalar"},
			{Key: vectorKey, Label: "SIMD", Repetition: 1, NanosecondsPerOp: 5, SourceBenchmark: "simd"},
			{Key: vectorKey, Label: "SIMD", Repetition: 2, NanosecondsPerOp: 6, SourceBenchmark: "simd"},
			{Key: vectorKey, Label: "SIMD", Repetition: 3, NanosecondsPerOp: 7, SourceBenchmark: "simd"},
		}},
	}
	got, err := InterpretCapture(capture)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Series) != 2 || len(got.Comparisons) != 1 {
		t.Fatalf("facts=%+v comparisons=%+v", got.Series, got.Comparisons)
	}
	if got.Series[0].MeanNS != 11 || !got.Series[0].SampleStdDev.Available || got.Series[0].SampleStdDev.Value != 1 || got.Series[0].IQR.Available {
		t.Fatalf("descriptive statistics=%+v", got.Series[0])
	}
	comparison := got.Comparisons[0]
	if comparison.TimeChangePercent != (6.0-11.0)/11.0*100 || comparison.SpeedupRatio != 11.0/6.0 {
		t.Fatalf("comparison=%+v", comparison)
	}
	if !strings.Contains(got.MeasurementNotice, "not a browser") || len(got.Claims.Forbidden) < 5 {
		t.Fatalf("claim boundaries missing: %+v", got)
	}
	if !slices.ContainsFunc(got.Caveats, func(c Caveat) bool { return c.Code == "fixed_iterations" }) {
		t.Fatalf("fixed-iteration caveat missing: %+v", got.Caveats)
	}
	if !slices.ContainsFunc(got.Caveats, func(c Caveat) bool { return c.Code == "parallel_vs_serial" }) {
		t.Fatalf("parallel caveat missing: %+v", got.Caveats)
	}
}

func TestInterpretCaptureComputesIQRWithEnoughSamples(t *testing.T) {
	key := result.MeasurementKey{RunID: "run", EnvironmentID: "env", BackendID: "backend", HarnessID: "harness", Case: result.Case{OperationID: catalog.Add, TypeID: catalog.Float32, StrategyID: catalog.Scalar, Elements: 8}}
	observations := make([]result.Observation, 0, 4)
	for index, value := range []float64{1, 2, 3, 4} {
		observations = append(observations, result.Observation{Key: key, Label: "Scalar", Repetition: index + 1, NanosecondsPerOp: value, SourceBenchmark: "scalar"})
	}
	capture := evidence.Capture{Schema: evidence.Schema, Status: "complete", Group: "addition", Run: result.Run{ID: "run", Observations: observations}}
	got, err := InterpretCapture(capture)
	if err != nil {
		t.Fatal(err)
	}
	series := got.Series[0]
	if series.MeanNS != 2.5 || !series.IQR.Available || series.IQR.Value != 2 || !series.SampleStdDev.Available || math.Abs(series.SampleStdDev.Value-math.Sqrt(5.0/3.0)) > 1e-12 {
		t.Fatalf("series=%+v", series)
	}
}

func TestInterpretFailedCheckpointAndRejectEmpty(t *testing.T) {
	key := result.MeasurementKey{Case: result.Case{OperationID: catalog.Add, StrategyID: catalog.Scalar, Elements: 8}}
	capture := evidence.Capture{Schema: evidence.Schema, Status: "failed", Group: "addition", Run: result.Run{ID: "partial", Observations: []result.Observation{{Key: key, Label: "Scalar", NanosecondsPerOp: 10, SourceBenchmark: "scalar"}}}}
	got, err := InterpretCapture(capture)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(got.Caveats, func(c Caveat) bool { return c.Code == "incomplete_run" }) {
		t.Fatalf("incomplete caveat missing: %+v", got.Caveats)
	}
	capture.Run.Observations = nil
	if _, err := InterpretCapture(capture); err == nil {
		t.Fatal("empty capture accepted")
	}
}
