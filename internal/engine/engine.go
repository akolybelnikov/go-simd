// Package engine runs registered kernels using a compiled benchmark adapter.
package engine

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

const BackendID = "testing-benchmark/v1"
const HarnessID = "standalone/v1"

type Config struct {
	Count     int    `json:"count"`
	Benchtime string `json:"benchtime"`
	Bench     string `json:"bench"`
}

func (c Config) Validate() error {
	if c.Count < 1 {
		return fmt.Errorf("count must be positive")
	}
	if strings.HasSuffix(c.Benchtime, "x") {
		n, err := strconv.ParseInt(strings.TrimSuffix(c.Benchtime, "x"), 10, strconv.IntSize)
		if err != nil || n < 1 {
			return fmt.Errorf("invalid benchtime %q: use a positive duration or iteration count Nx", c.Benchtime)
		}
	} else if d, err := time.ParseDuration(c.Benchtime); err != nil || d <= 0 {
		return fmt.Errorf("invalid benchtime %q: use a positive duration or iteration count Nx", c.Benchtime)
	}
	_, err := compileSelection(c.Bench)
	return err
}

type Case struct {
	Operation      catalog.Operation
	Implementation kernels.Implementation
	Elements       int
	Bytes          int64
	Name           string
}

type Plan struct {
	Config Config
	Cases  []Case
}

// Prepare resolves every selection and binding before allocating large fixtures.
func Prepare(group string, c Config, width int) (Plan, error) {
	if err := c.Validate(); err != nil {
		return Plan{}, err
	}
	op, ok := catalog.ByGroup(group)
	if !ok {
		return Plan{}, fmt.Errorf("unknown group %q", group)
	}
	bound, err := kernels.Lookup(op.ID, op.TypeID)
	if err != nil {
		return Plan{}, err
	}
	match, _ := compileSelection(c.Bench)
	plan := Plan{Config: c}
	for _, n := range benchutil.BenchmarkSizes(width) {
		bytes, err := op.LogicalBytes(n)
		if err != nil {
			return Plan{}, err
		}
		for _, impl := range bound.Implementations {
			name := fmt.Sprintf("%s/n=%d/%s", op.BenchmarkName, n, impl.Descriptor.Label)
			if match(name) {
				plan.Cases = append(plan.Cases, Case{op, impl, n, bytes, name})
			}
		}
	}
	return plan, nil
}

type Prepared struct {
	Run    func() error
	Output []float32
	Bytes  int64
}
type Sample struct {
	Iterations          int
	Duration            time.Duration
	MemAllocs, MemBytes uint64
}
type Backend interface {
	Measure(Prepared, string) (Sample, error)
}

// Execute retains successful samples on a later unexpected kernel/backend error.
// Fixtures are reused per size, with no implicit warmup or output checksum.
func Execute(plan Plan, run result.Run, backend Backend) (result.Run, error) {
	if len(plan.Cases) == 0 {
		return run, fmt.Errorf("no benchmark measurements found")
	}
	procs := runtime.GOMAXPROCS(0)
	var a, b, out []float32
	previous := -1
	fingerprint := ""
	for _, c := range plan.Cases {
		if c.Elements != previous {
			a, b, out = input.MakeFloat32Slices(c.Elements)
			fingerprint = input.FingerprintFloat32(a, b)
			previous = c.Elements
		}
		// Native developer benchmarks retain the same callback -> typed function
		// pointer shape. The standalone harness has its own versioned identity.
		prepared := Prepared{Run: func() error { return c.Implementation.Run(a, b, out) }, Output: out, Bytes: c.Bytes}
		workers, known := c.Implementation.Descriptor.WorkerCount(c.Elements, procs)
		key := result.MeasurementKey{
			RunID: run.ID, EnvironmentID: run.Environment.ID,
			BackendID: run.BackendID, HarnessID: run.HarnessID,
			GOMAXPROCS: result.OptionalInt{Value: procs, Known: true},
			Case: result.Case{
				OperationID: c.Operation.ID, TypeID: c.Operation.TypeID,
				StrategyID: c.Implementation.Descriptor.ID, ArithmeticID: c.Operation.ArithmeticID,
				Elements:      c.Elements,
				Input:         result.InputIdentity{GeneratorID: input.GeneratedFloat32ID, SHA256: fingerprint},
				Workers:       result.OptionalInt{Value: workers, Known: known},
				ExecutionMode: c.Operation.ExecutionMode,
			},
		}
		for repetition := 1; repetition <= plan.Config.Count; repetition++ {
			sample, err := backend.Measure(prepared, plan.Config.Benchtime)
			if err != nil {
				return run, fmt.Errorf("%s repetition %d: %w", c.Name, repetition, err)
			}
			if sample.Iterations < 1 || sample.Duration < 0 {
				return run, fmt.Errorf("invalid measurement for %s", c.Name)
			}
			n := float64(sample.Iterations)
			duration := sample.Duration.Nanoseconds()
			throughput := result.Metric{}
			if duration > 0 {
				throughput = result.Metric{Value: float64(c.Bytes) * n * 1000 / float64(duration), Available: true}
			}
			run.Observations = append(run.Observations, result.Observation{
				Key: key, Label: c.Implementation.Descriptor.Label, Repetition: repetition,
				Iterations: uint64(sample.Iterations), DurationNS: &duration,
				NanosecondsPerOp: float64(duration) / n,
				BytesPerOp:       result.Metric{Value: float64(sample.MemBytes) / n, Available: true},
				AllocsPerOp:      result.Metric{Value: float64(sample.MemAllocs) / n, Available: true},
				MegabytesPerSec:  throughput, SourceBenchmark: fmt.Sprintf("%s-%d", c.Name, procs),
			})
		}
	}
	return run, nil
}
