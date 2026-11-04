// Package analysis converts measurement evidence into deterministic facts and
// claim boundaries. It contains no model-generated interpretation.
package analysis

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/result"
)

const Schema = "go-simd.interpretation/v1"

type EnvironmentFact struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	CPU          string `json:"cpu"`
	VectorBits   int    `json:"vector_bits"`
	Float32Lanes int    `json:"float32_lanes"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	Emulated     bool   `json:"emulated"`
}

type SeriesFact struct {
	FactID        string             `json:"fact_id"`
	OperationID   string             `json:"operation_id"`
	StrategyID    string             `json:"strategy_id"`
	Strategy      string             `json:"strategy"`
	ExecutionMode string             `json:"execution_mode"`
	Elements      int                `json:"elements"`
	Samples       int                `json:"samples"`
	MedianNS      float64            `json:"median_ns_per_op"`
	MeanNS        float64            `json:"mean_ns_per_op"`
	MinimumNS     float64            `json:"minimum_ns_per_op"`
	MaximumNS     float64            `json:"maximum_ns_per_op"`
	SampleStdDev  OptionalFactMetric `json:"sample_standard_deviation_ns_per_op"`
	IQR           OptionalFactMetric `json:"interquartile_range_ns_per_op"`
}

type OptionalFactMetric struct {
	Value     float64 `json:"value,omitempty"`
	Available bool    `json:"available"`
	Reason    string  `json:"reason,omitempty"`
}

type ComparisonFact struct {
	FactID            string  `json:"fact_id"`
	BaselineStrategy  string  `json:"baseline_strategy_id"`
	CandidateStrategy string  `json:"candidate_strategy_id"`
	Elements          int     `json:"elements"`
	TimeChangePercent float64 `json:"time_change_percent"`
	SpeedupRatio      float64 `json:"speedup_ratio"`
}

type Caveat struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type ClaimPolicy struct {
	Allowed   []string `json:"allowed"`
	Forbidden []string `json:"forbidden"`
}

type Interpretation struct {
	Schema               string           `json:"schema"`
	RunID                string           `json:"run_id"`
	Status               string           `json:"status"`
	MeasuringEnvironment EnvironmentFact  `json:"measuring_environment"`
	MeasurementNotice    string           `json:"measurement_notice"`
	Series               []SeriesFact     `json:"series"`
	Comparisons          []ComparisonFact `json:"comparisons,omitempty"`
	Caveats              []Caveat         `json:"caveats"`
	Claims               ClaimPolicy      `json:"claims"`
}

var allowedClaims = []string{
	"exact measuring-environment identity reported by the capture",
	"observed sample count, minimum, median, and maximum",
	"descriptive within-run time change and speedup for compatible measured series",
	"explicit fixture, parameter, backend, harness, and execution identities present in evidence",
}

var forbiddenClaims = []string{
	"statistical significance or confidence not computed by the tool",
	"causal hardware, compiler, memory, or pipeline bottleneck attribution",
	"claims about the browser or client machine from server measurements",
	"extrapolated production, monetary, energy, or capacity savings",
	"general workload recommendations beyond the measured operation, input, sizes, and runtime",
}

// InterpretCapture emits only directly traceable descriptive facts. Failed
// captures remain usable when they contain completed checkpoint observations.
func InterpretCapture(capture evidence.Capture) (Interpretation, error) {
	if capture.Schema != evidence.Schema {
		return Interpretation{}, fmt.Errorf("unsupported measurement schema %q", capture.Schema)
	}
	if capture.Status != "complete" && capture.Status != "failed" {
		return Interpretation{}, fmt.Errorf("invalid measurement status %q", capture.Status)
	}
	if len(capture.Run.Observations) == 0 {
		return Interpretation{}, fmt.Errorf("no completed observations to interpret")
	}
	summaries, err := result.Summarize(capture.Run.Observations)
	if err != nil {
		return Interpretation{}, err
	}
	slices.SortFunc(summaries, func(a, b result.Summary) int {
		if a.Key.Case.Elements != b.Key.Case.Elements {
			return a.Key.Case.Elements - b.Key.Case.Elements
		}
		return strings.Compare(a.Key.Case.StrategyID, b.Key.Case.StrategyID)
	})
	output := Interpretation{
		Schema: Schema, RunID: capture.Run.ID, Status: capture.Status,
		MeasuringEnvironment: EnvironmentFact{
			GoVersion: capture.Machine.GoVersion, OS: capture.Machine.OS, Arch: capture.Machine.Arch,
			CPU: capture.Machine.CPU, VectorBits: capture.Machine.Bits, Float32Lanes: capture.Machine.Lanes,
			GOMAXPROCS: capture.Machine.Procs, Emulated: capture.Machine.Emulated,
		},
		MeasurementNotice: "These measurements describe the measuring process and host recorded above, not a browser or remote client device.",
		Claims:            ClaimPolicy{Allowed: slices.Clone(allowedClaims), Forbidden: slices.Clone(forbiddenClaims)},
	}
	for index, summary := range summaries {
		stats := describeSeries(capture.Run.Observations, summary.Key)
		output.Series = append(output.Series, SeriesFact{
			FactID: "series-" + strconv.Itoa(index+1), OperationID: summary.Key.Case.OperationID,
			StrategyID: summary.Key.Case.StrategyID, Strategy: summary.Label, Elements: summary.Key.Case.Elements,
			Samples: summary.Samples, MedianNS: summary.Median, MeanNS: stats.mean,
			MinimumNS: stats.minimum, MaximumNS: stats.maximum,
			SampleStdDev: stats.standardDeviation, IQR: stats.iqr,
			ExecutionMode: summary.Key.Case.ExecutionMode,
		})
		if summary.Samples < 3 {
			addCaveat(&output, "low_sample_count", "At least one series has fewer than three observations; variability evidence is very limited.")
		}
	}
	comparison, err := result.Compare(capture.Group, summaries)
	if err != nil {
		return Interpretation{}, fmt.Errorf("build compatible comparison: %w", err)
	}
	for _, row := range comparison.Rows {
		baseline := metricFor(comparison, row, catalog.Scalar)
		if !baseline.Available || baseline.Value <= 0 {
			continue
		}
		for index, column := range comparison.Columns {
			candidate := row.Cells[index].NanosecondsPerOp
			if column.StrategyID == catalog.Scalar || !candidate.Available || candidate.Value <= 0 {
				continue
			}
			change := (candidate.Value - baseline.Value) / baseline.Value * 100
			speedup := baseline.Value / candidate.Value
			if !finite(change) || !finite(speedup) {
				return Interpretation{}, fmt.Errorf("non-finite descriptive comparison")
			}
			output.Comparisons = append(output.Comparisons, ComparisonFact{
				FactID: "comparison-" + strconv.Itoa(len(output.Comparisons)+1), Elements: row.Elements,
				BaselineStrategy: catalog.Scalar, CandidateStrategy: column.StrategyID,
				TimeChangePercent: change, SpeedupRatio: speedup,
			})
		}
	}
	if strings.HasSuffix(capture.Config.Benchtime, "x") {
		addCaveat(&output, "fixed_iterations", "Fixed iteration measurements may be too short for stable performance evidence.")
	}
	if capture.Status == "failed" {
		addCaveat(&output, "incomplete_run", "The run failed or was interrupted; only completed checkpoint observations are included.")
	}
	if capture.Machine.Emulated {
		addCaveat(&output, "emulated_simd", "The capture reports an emulated SIMD implementation.")
	}
	addCaveat(&output, "host_contention_unmeasured", "Host contention and background load were not measured; execution was serialized only within this lab service.")
	if mixesParallelAndSerial(capture.Run.Observations) {
		addCaveat(&output, "parallel_vs_serial", "The comparison includes concurrency effects as well as implementation differences.")
	}
	return output, nil
}

type seriesStatistics struct {
	mean, minimum, maximum float64
	standardDeviation      OptionalFactMetric
	iqr                    OptionalFactMetric
}

func describeSeries(observations []result.Observation, key result.MeasurementKey) seriesStatistics {
	values := make([]float64, 0)
	for _, observation := range observations {
		if observation.Key == key {
			values = append(values, observation.NanosecondsPerOp)
		}
	}
	slices.Sort(values)
	stats := seriesStatistics{minimum: values[0], maximum: values[len(values)-1]}
	for _, value := range values {
		stats.mean += value
	}
	stats.mean /= float64(len(values))
	if len(values) < 2 {
		stats.standardDeviation.Reason = "at least two observations are required"
	} else {
		var squared float64
		for _, value := range values {
			difference := value - stats.mean
			squared += difference * difference
		}
		stats.standardDeviation = OptionalFactMetric{Value: math.Sqrt(squared / float64(len(values)-1)), Available: true}
	}
	if len(values) < 4 {
		stats.iqr.Reason = "at least four observations are required"
	} else {
		middle := len(values) / 2
		lower, upper := values[:middle], values[(len(values)+1)/2:]
		stats.iqr = OptionalFactMetric{Value: median(upper) - median(lower), Available: true}
	}
	return stats
}

func median(values []float64) float64 {
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}

func metricFor(comparison result.Comparison, row result.Row, strategy string) result.Metric {
	for index, column := range comparison.Columns {
		if column.StrategyID == strategy {
			return row.Cells[index].NanosecondsPerOp
		}
	}
	return result.Metric{}
}

func addCaveat(output *Interpretation, code, message string) {
	if !slices.ContainsFunc(output.Caveats, func(c Caveat) bool { return c.Code == code }) {
		output.Caveats = append(output.Caveats, Caveat{Code: code, Message: message})
	}
}

func mixesParallelAndSerial(observations []result.Observation) bool {
	hasSerial, hasParallel := false, false
	for _, observation := range observations {
		workers := observation.Key.Case.Workers
		if workers.Known && workers.Value > 1 {
			hasParallel = true
		} else {
			hasSerial = true
		}
	}
	return hasSerial && hasParallel
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
