// Package engine runs registered kernels using a compiled benchmark adapter.
package engine

import (
	"crypto/sha256"
	"fmt"
	"math"
	"runtime"
	"simd"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
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
	Input  input.Options
	Config Config
	Cases  []Case
}

// Prepare resolves every selection and binding before allocating large fixtures.
func Prepare(group string, c Config, width int) (Plan, error) {
	return PrepareWithOptions(group, c, width, input.DefaultOptions())
}

func PrepareWithOptions(group string, c Config, width int, options input.Options) (Plan, error) {
	if err := c.Validate(); err != nil {
		return Plan{}, err
	}
	op, ok := catalog.ByGroup(group)
	if !ok {
		return Plan{}, fmt.Errorf("unknown group %q", group)
	}
	if err := options.ValidateFor(op); err != nil {
		return Plan{}, err
	}
	bound, err := kernels.Lookup(op.ID, op.TypeID)
	if err != nil {
		return Plan{}, err
	}
	match, _ := compileSelection(c.Bench)
	plan := Plan{Config: c, Input: options}
	for _, n := range benchutil.OperationSizes(op, width) {
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
	ByteOutput    []uint8
	ReduceChecked func() (float32, error)
	Reduce        func() float32
	Scalar        *float32
	Run           func() error
	Output        []float32
	Bytes         int64
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
	var byteOut []uint8
	var scalar float32
	previous := -1
	fingerprint := ""
	for _, c := range plan.Cases {
		if c.Elements != previous {
			a, b, out = input.MakeForOperation(c.Operation, c.Elements, plan.Input)
			fingerprint = input.FingerprintFloat32(a, b)
			previous = c.Elements
			if c.Operation.ByteResult() {
				length, err := c.Operation.OutputElements(c.Elements)
				if err != nil {
					return run, err
				}
				byteOut = make([]uint8, length)
			}
		}
		// Native developer benchmarks retain the same callback -> typed function
		// pointer shape. The standalone harness has its own versioned identity.
		prepared := Prepared{Run: func() error { return c.Implementation.Run(a, b, out) }, Output: out, Bytes: c.Bytes}
		if c.Operation.ByteResult() {
			prepared.Run = c.Implementation.BindUint8(a, byteOut, plan.Input)
			prepared.ByteOutput = byteOut
		} else if c.Operation.Parameterized() {
			prepared.Run = c.Implementation.BindConfigured(a, b, out, plan.Input)
		}
		if c.Operation.ScalarResult() {
			prepared.Run = nil
			if c.Implementation.BindCheckedScalar != nil {
				prepared.ReduceChecked = c.Implementation.BindCheckedScalar(a, b)
			} else {
				prepared.Reduce = c.Implementation.BindScalar(a)
			}
			prepared.Scalar = &scalar
		}
		workers, known := c.Implementation.Descriptor.WorkerCount(c.Elements, procs)
		key := result.MeasurementKey{
			RunID: run.ID, EnvironmentID: run.Environment.ID,
			BackendID: run.BackendID, HarnessID: run.HarnessID,
			GOMAXPROCS: result.OptionalInt{Value: procs, Known: true},
			Case: result.Case{
				OperationID: c.Operation.ID, TypeID: c.Operation.TypeID,
				StrategyID: c.Implementation.Descriptor.ID, ArithmeticID: c.Operation.ArithmeticID,
				Elements:      c.Elements,
				Parameters:    plan.Input.Identity(c.Operation),
				Input:         result.InputIdentity{GeneratorID: plan.Input.Generator(c.Operation), SHA256: fingerprint},
				Workers:       result.OptionalInt{Value: workers, Known: known},
				ExecutionMode: c.Operation.ExecutionMode,
			},
		}
		if c.Operation.ScalarResult() {
			var vector simd.Float32s
			fold, err := result.ReductionFold(c.Operation.ID, c.Implementation.Descriptor.ID, vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Reduction = fold
		}
		if c.Operation.ByteResult() {
			var vector simd.Float32s
			identity, err := result.NewPredicateIdentity(vector.Len())
			if c.Operation.ID == catalog.PredicateGTPacked {
				identity, err = result.NewPackedPredicateIdentity(vector.Len())
			}
			if err != nil {
				return run, err
			}
			key.Case.Predicate = identity
		}
		if c.Operation.ID == catalog.AffineFMA {
			var vector simd.Float32s
			identity, err := result.NewFMAIdentity(c.Implementation.Descriptor.ID, fmacap.Path(), vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.FMA = identity
		}
		if c.Operation.ID == catalog.DotFMA {
			var vector simd.Float32s
			identity, err := result.NewDotFMAIdentity(c.Implementation.Descriptor.ID, fmacap.Path(), vector.Len(), c.Elements)
			if err != nil {
				return run, err
			}
			key.Case.FMA = identity
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
			if c.Operation.ScalarResult() {
				bits := math.Float32bits(scalar)
				run.Observations[len(run.Observations)-1].ScalarResultBits = &bits
			}
			if c.Operation.ByteResult() {
				digest := sha256.Sum256(byteOut)
				run.Observations[len(run.Observations)-1].ByteResultSHA256 = fmt.Sprintf("%x", digest)
			}
			if c.Operation.ID == catalog.AffineFMA {
				bits := math.Float32bits(out[len(out)-1])
				run.Observations[len(run.Observations)-1].ArrayResultLastBits = &bits
			}
		}
	}
	return run, nil
}
