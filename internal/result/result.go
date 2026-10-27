// Package result retains measurement identity and computes pure summaries.
// It has no dependency on SIMD, testing, command execution, or rendering.
package result

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// Optional values distinguish an unavailable metric from a measured zero.
type Metric struct {
	Value     float64
	Available bool
}

type OptionalInt struct {
	Value int
	Known bool
}

type InputIdentity struct {
	GeneratorID string
	SHA256      string // Empty means unknown, never an inferred match.
}

// Case is a comparable logical identity, independent of display labels.
type Case struct {
	OperationID, TypeID, StrategyID, ArithmeticID string
	Parameters                                    string
	Elements                                      int
	Input                                         InputIdentity
	Workers                                       OptionalInt
	ExecutionMode                                 string
}

type MeasurementKey struct {
	RunID, EnvironmentID, BackendID, HarnessID string
	GOMAXPROCS                                 OptionalInt
	Case                                       Case
}

type Observation struct {
	Key              MeasurementKey
	Label            string
	Repetition       int // One-based within the complete measurement key.
	Iterations       uint64
	NanosecondsPerOp float64
	DurationNS       *int64 // Legacy rounded ns/op cannot recover exact duration.
	BytesPerOp       Metric
	AllocsPerOp      Metric
	MegabytesPerSec  Metric
	SourceBenchmark  string
}

type Environment struct {
	ID, GOOS, GOARCH, CPU string // Only facts reported by the measurement process.
}

type Run struct {
	ID           string
	Environment  Environment
	BackendID    string
	HarnessID    string
	SourceFormat string
	Metadata     string // Original readable metadata, not a substitute for identity.
	Observations []Observation
}

type Summary struct {
	Key     MeasurementKey
	Label   string
	Median  float64
	Samples int
}

// Summarize preserves independent series and never mutates the observations.
func Summarize(observations []Observation) ([]Summary, error) {
	values := make(map[MeasurementKey][]float64)
	labels := make(map[MeasurementKey]string)
	var keys []MeasurementKey
	for _, o := range observations {
		if !finite(o.NanosecondsPerOp) || o.NanosecondsPerOp < 0 {
			return nil, fmt.Errorf("invalid ns/op for %s", o.SourceBenchmark)
		}
		if _, ok := values[o.Key]; !ok {
			keys = append(keys, o.Key)
			labels[o.Key] = o.Label
		} else if labels[o.Key] != o.Label {
			return nil, fmt.Errorf("conflicting labels for strategy %q", o.Key.Case.StrategyID)
		}
		values[o.Key] = append(values[o.Key], o.NanosecondsPerOp)
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no benchmark measurements found")
	}
	summaries := make([]Summary, 0, len(keys))
	for _, key := range keys {
		v := values[key]
		slices.Sort(v)
		n := len(v)
		median := v[n/2]
		if n%2 == 0 {
			median = (v[n/2-1] + median) / 2
		}
		summaries = append(summaries, Summary{key, labels[key], median, n})
	}
	return summaries, nil
}

type Column struct {
	StrategyID string
	Label      string
	Baseline   bool
}

type Cell struct {
	NanosecondsPerOp Metric
	ChangePercent    Metric
	Samples          int
}

type Row struct {
	Elements int
	Cells    []Cell // Matches Comparison.Columns.
}

type Comparison struct {
	Title   string
	Columns []Column
	Rows    []Row
	// Legacy tables compare every strategy (including parallel) to serial scalar.
	// This is not evidence attributing concurrency gains to vectorization.
	BaselinePolicy string
}

// comparisonContext omits strategy/workers for the legacy serial-scalar policy,
// and size/fingerprint because the report intentionally sweeps input sizes.
type comparisonContext struct {
	RunID, EnvironmentID, BackendID, HarnessID string
	GOMAXPROCS                                 OptionalInt
	OperationID, TypeID, ArithmeticID          string
	Parameters, GeneratorID, ExecutionMode     string
}

func contextFor(k MeasurementKey) comparisonContext {
	c := k.Case
	return comparisonContext{k.RunID, k.EnvironmentID, k.BackendID, k.HarnessID,
		k.GOMAXPROCS, c.OperationID, c.TypeID, c.ArithmeticID,
		c.Parameters, c.Input.GeneratorID, c.ExecutionMode}
}

// Compare builds one compatible size sweep. Mixed contexts and ambiguous cases
// fail rather than silently collapsing distinct operations, inputs, or workers.
func Compare(title string, summaries []Summary) (Comparison, error) {
	if len(summaries) == 0 {
		return Comparison{}, fmt.Errorf("no benchmark measurements found")
	}
	context := contextFor(summaries[0].Key)
	columns := map[string]Column{}
	bySize := map[int]map[string]Summary{}
	fingerprints := map[int]string{}
	for _, s := range summaries {
		if contextFor(s.Key) != context {
			return Comparison{}, fmt.Errorf("incompatible measurement contexts in comparison")
		}
		if !finite(s.Median) || s.Median < 0 || s.Samples < 1 {
			return Comparison{}, fmt.Errorf("invalid summary for strategy %q", s.Key.Case.StrategyID)
		}
		c := s.Key.Case
		column := Column{c.StrategyID, s.Label, c.StrategyID == "scalar"}
		if previous, ok := columns[c.StrategyID]; ok && previous != column {
			return Comparison{}, fmt.Errorf("conflicting strategy labels")
		}
		columns[c.StrategyID] = column
		if bySize[c.Elements] == nil {
			bySize[c.Elements] = map[string]Summary{}
			fingerprints[c.Elements] = c.Input.SHA256
		} else if fingerprints[c.Elements] != c.Input.SHA256 {
			return Comparison{}, fmt.Errorf("incompatible inputs at size %d", c.Elements)
		}
		if _, exists := bySize[c.Elements][c.StrategyID]; exists {
			return Comparison{}, fmt.Errorf("ambiguous strategy %q at size %d (check worker configuration)", c.StrategyID, c.Elements)
		}
		bySize[c.Elements][c.StrategyID] = s
	}
	table := Comparison{Title: title, BaselinePolicy: "legacy-serial-scalar/v1"}
	for _, column := range columns {
		table.Columns = append(table.Columns, column)
	}
	slices.SortFunc(table.Columns, func(a, b Column) int {
		order := []string{"Scalar", "SIMD", "SIMDPartial", "SIMDTail", "SIMDParallel"}
		ai, bi := slices.Index(order, a.Label), slices.Index(order, b.Label)
		if ai < 0 {
			ai = len(order)
		}
		if bi < 0 {
			bi = len(order)
		}
		if ai != bi {
			return ai - bi
		}
		return strings.Compare(a.Label, b.Label)
	})
	sizes := make([]int, 0, len(bySize))
	for size := range bySize {
		sizes = append(sizes, size)
	}
	slices.Sort(sizes)
	for _, size := range sizes {
		row := Row{Elements: size}
		baseline, hasBaseline := bySize[size]["scalar"]
		for _, column := range table.Columns {
			cell := Cell{}
			if s, ok := bySize[size][column.StrategyID]; ok {
				cell.NanosecondsPerOp = Metric{s.Median, true}
				cell.Samples = s.Samples
				if !column.Baseline && hasBaseline && baseline.Median > 0 {
					change := (s.Median - baseline.Median) / baseline.Median * 100
					if !finite(change) {
						return Comparison{}, fmt.Errorf("non-finite percentage change")
					}
					cell.ChangePercent = Metric{change, true}
				}
			}
			row.Cells = append(row.Cells, cell)
		}
		table.Rows = append(table.Rows, row)
	}
	return table, nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
