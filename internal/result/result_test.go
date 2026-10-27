package result

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func fixtureObservations(t *testing.T) []Observation {
	t.Helper()
	run, err := ReadLegacy(strings.NewReader(
		"BenchmarkAddition/n=8/Scalar-4 1 10 ns/op\n"+
			"BenchmarkAddition/n=8/Scalar-4 1 30 ns/op\n"+
			"BenchmarkAddition/n=8/SIMDPartial-4 1 8.125 ns/op\n"+
			"BenchmarkAddition/n=8/SIMDPartial-4 1 8.375 ns/op\n"+
			"BenchmarkAddition/n=8/SIMDPartial-4 1 8.25 ns/op\n"), LegacyOptions{RunID: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	return run.Observations
}

func TestSummariesAndRatioOfMedians(t *testing.T) {
	observations := fixtureObservations(t)
	before := append([]Observation(nil), observations...)
	summaries, err := Summarize(observations)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observations, before) {
		t.Fatal("aggregation mutated source observations")
	}
	if len(summaries) != 2 || summaries[0].Median != 20 || summaries[0].Samples != 2 || summaries[1].Median != 8.25 || summaries[1].Samples != 3 {
		t.Fatalf("incorrect odd/even medians: %+v", summaries)
	}
	table, err := Compare("addition", summaries)
	if err != nil {
		t.Fatal(err)
	}
	if table.Rows[0].Cells[1].ChangePercent != (Metric{-58.75, true}) || table.BaselinePolicy != "legacy-serial-scalar/v1" {
		t.Fatalf("change must use unrounded median ratio: %+v", table)
	}
}

func TestMeasurementIdentitiesNeverCoalesce(t *testing.T) {
	base := fixtureObservations(t)[0]
	for name, change := range map[string]func(*Observation){
		"operation":   func(o *Observation) { o.Key.Case.OperationID = "other" },
		"type":        func(o *Observation) { o.Key.Case.TypeID = "float64" },
		"semantics":   func(o *Observation) { o.Key.Case.ArithmeticID = "fused" },
		"parameters":  func(o *Observation) { o.Key.Case.Parameters = "changed" },
		"input":       func(o *Observation) { o.Key.Case.Input.SHA256 = "different" },
		"generator":   func(o *Observation) { o.Key.Case.Input.GeneratorID = "other" },
		"workers":     func(o *Observation) { o.Key.Case.Workers = OptionalInt{2, true} },
		"mode":        func(o *Observation) { o.Key.Case.ExecutionMode = "end-to-end" },
		"run":         func(o *Observation) { o.Key.RunID = "other" },
		"environment": func(o *Observation) { o.Key.EnvironmentID = "other" },
		"backend":     func(o *Observation) { o.Key.BackendID = "other/v1" },
		"harness":     func(o *Observation) { o.Key.HarnessID = "other/v1" },
		"GOMAXPROCS":  func(o *Observation) { o.Key.GOMAXPROCS.Value++ },
	} {
		t.Run(name, func(t *testing.T) {
			other := base
			change(&other)
			other.NanosecondsPerOp = 100
			summaries, err := Summarize([]Observation{base, other})
			if err != nil {
				t.Fatal(err)
			}
			if len(summaries) != 2 {
				t.Fatal("independent cases pooled")
			}
			if _, err := Compare("mixed", summaries); err == nil {
				t.Fatal("incompatible or ambiguous cases entered one table")
			}
		})
	}
}

func TestParallelStillComparesToSerialScalar(t *testing.T) {
	run, err := ReadLegacy(strings.NewReader("BenchmarkAddition/n=8/Scalar-4 1 10 ns/op\nBenchmarkAddition/n=8/SIMDParallel-4 1 5 ns/op\n"), LegacyOptions{RunID: "legacy"})
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := Summarize(run.Observations)
	if err != nil {
		t.Fatal(err)
	}
	table, err := Compare("legacy", summaries)
	if err != nil {
		t.Fatal(err)
	}
	if table.Rows[0].Cells[1].ChangePercent != (Metric{-50, true}) {
		t.Fatal("legacy serial baseline policy changed")
	}
}

func TestInvalidObservationsAndEmptySummaries(t *testing.T) {
	base := fixtureObservations(t)[0]
	for _, value := range []float64{-1, math.Inf(1), math.NaN()} {
		base.NanosecondsPerOp = value
		if _, err := Summarize([]Observation{base}); err == nil {
			t.Errorf("accepted invalid timing %v", value)
		}
	}
	if _, err := Summarize(nil); err == nil {
		t.Fatal("accepted empty observations")
	}
	if _, err := Compare("empty", nil); err == nil {
		t.Fatal("accepted empty summaries")
	}
}
