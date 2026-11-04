// Package engine runs registered kernels using a compiled benchmark adapter.
package engine

import (
	"context"
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
	Float64Output           []float64
	RetainOutputAfterTiming bool
	ByteOutput              []uint8
	ReduceChecked           func() (float32, error)
	Reduce                  func() float32
	Scalar                  *float32
	Run                     func() error
	Output                  []float32
	Bytes                   int64
}
type Sample struct {
	Iterations          int
	Duration            time.Duration
	MemAllocs, MemBytes uint64
}
type Backend interface {
	Measure(Prepared, string) (Sample, error)
}

// ContextBackend lets a backend stop setup or external work when cancellation is
// requested. Timed kernel loops are not instrumented with context checks.
type ContextBackend interface {
	MeasureContext(context.Context, Prepared, string) (Sample, error)
}

// Execute retains successful samples on a later unexpected kernel/backend error.
// Fixtures are reused per size; declared output digests are computed after samples.
func Execute(plan Plan, run result.Run, backend Backend) (result.Run, error) {
	return ExecuteContext(context.Background(), plan, run, backend)
}

// ExecuteContext checks cancellation between measurement repetitions and retains
// every successfully completed sample. A cancellation during one in-process
// testing.Benchmark takes effect after that sample; isolated workers can provide
// hard cancellation without adding checks to timed kernel bodies.
func ExecuteContext(ctx context.Context, plan Plan, run result.Run, backend Backend) (result.Run, error) {
	return ExecuteObserved(ctx, plan, run, backend, nil)
}

// ExecuteObserved publishes an immutable snapshot after every completed
// observation. Isolated workers use it to checkpoint recoverable progress.
func ExecuteObserved(ctx context.Context, plan Plan, run result.Run, backend Backend, observe func(result.Run) error) (result.Run, error) {
	if ctx == nil {
		return run, fmt.Errorf("nil execution context")
	}
	if backend == nil {
		return run, fmt.Errorf("nil measurement backend")
	}
	if len(plan.Cases) == 0 {
		return run, fmt.Errorf("no benchmark measurements found")
	}
	procs := runtime.GOMAXPROCS(0)
	var a, b, third, out []float32
	var doubleA, doubleB, doubleOut []float64
	var byteA, byteB, byteOut []uint8
	var scalar float32
	previous := -1
	fingerprint := ""
	for _, c := range plan.Cases {
		if err := ctx.Err(); err != nil {
			return run, fmt.Errorf("before %s: %w", c.Name, err)
		}
		if c.Elements != previous {
			if c.Operation.ID == catalog.SaturatingAdd8 {
				byteA, byteB, byteOut = input.MakeSaturatingAdd8(c.Elements)
				fingerprint = input.FingerprintUint8(byteA, byteB)
			} else if c.Operation.ID == catalog.Add64 {
				doubleA, doubleB, doubleOut = input.MakeAddition64(c.Elements)
				fingerprint = input.FingerprintFloat64(doubleA, doubleB)
			} else if c.Operation.ID == catalog.MulAdd {
				a, b, third, out = input.MakeMulAdd(c.Elements, plan.Input)
				fingerprint = input.FingerprintTernaryFloat32(a, b, third)
			} else {
				a, b, out = input.MakeForOperation(c.Operation, c.Elements, plan.Input)
				fingerprint = input.FingerprintFloat32(a, b)
			}
			previous = c.Elements
			if c.Operation.ByteResult() && c.Operation.ID != catalog.SaturatingAdd8 {
				length, err := c.Operation.OutputElements(c.Elements)
				if err != nil {
					return run, err
				}
				byteOut = make([]uint8, length)
			}
		}
		// Original native benchmarks retain the same callback -> typed function
		// pointer shape. The standalone harness has its own versioned identity.
		prepared := Prepared{Run: func() error { return c.Implementation.Run(a, b, out) }, Output: out, Bytes: c.Bytes}
		if c.Operation.ID == catalog.SaturatingAdd8 {
			prepared.Run = c.Implementation.BindByteArrays(byteA, byteB, byteOut)
			prepared.Output = nil
			prepared.ByteOutput = byteOut
		} else if c.Operation.ID == catalog.Add64 {
			prepared.Run = c.Implementation.BindFloat64(doubleA, doubleB, doubleOut)
			prepared.Output = nil
			prepared.Float64Output = doubleOut
			prepared.RetainOutputAfterTiming = true
		} else if c.Operation.ID == catalog.MulAdd {
			prepared.Run = c.Implementation.BindTernary(a, b, third, out)
			prepared.RetainOutputAfterTiming = true
		} else if c.Operation.ID == catalog.Copy || c.Operation.ID == catalog.BitRoundTrip {
			prepared.Run = c.Implementation.BindUnary(a, out)
			prepared.RetainOutputAfterTiming = true
		} else if c.Operation.ID == catalog.Clamp {
			prepared.Run = c.Implementation.BindConfigured(a, b, out, plan.Input)
			prepared.RetainOutputAfterTiming = true
		} else if c.Operation.ByteResult() {
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
		if c.Operation.ByteResult() && c.Operation.TypeID == catalog.Float32 {
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
		if c.Operation.ID == catalog.BitRoundTrip {
			var vector simd.Float32s
			identity, err := result.NewVectorIdentity(vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Vector = identity
		}
		if c.Operation.ID == catalog.SaturatingAdd8 {
			var vector simd.Uint8s
			identity, err := result.NewVectorIdentity(vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Vector = identity
		}
		if c.Operation.ID == catalog.Add64 {
			var vector simd.Float64s
			identity, err := result.NewVectorIdentity(vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Vector = identity
		}
		if c.Operation.ID == catalog.Copy {
			var vector simd.Float32s
			identity, err := result.NewCopyIdentity(vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Copy = identity
		}
		if c.Operation.ID == catalog.Clamp {
			var vector simd.Float32s
			identity, err := result.NewClampIdentity(vector.Len())
			if err != nil {
				return run, err
			}
			key.Case.Clamp = identity
		}
		if c.Operation.ID == catalog.AffineFMA || c.Operation.ID == catalog.MulAdd {
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
			if err := ctx.Err(); err != nil {
				return run, fmt.Errorf("before %s repetition %d: %w", c.Name, repetition, err)
			}
			sample, err := measureContext(ctx, backend, prepared, plan.Config.Benchtime)
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
			if c.Operation.ID == catalog.Add64 {
				run.Observations[len(run.Observations)-1].ArrayResultSHA256 = input.FingerprintFloat64(doubleOut, nil)
			}
			if c.Operation.ID == catalog.Copy || c.Operation.ID == catalog.Clamp || c.Operation.ID == catalog.MulAdd || c.Operation.ID == catalog.BitRoundTrip {
				run.Observations[len(run.Observations)-1].ArrayResultSHA256 = input.FingerprintFloat32(out, nil)
			}
			if c.Operation.ID == catalog.AffineFMA {
				bits := math.Float32bits(out[len(out)-1])
				run.Observations[len(run.Observations)-1].ArrayResultLastBits = &bits
			}
			if observe != nil {
				snapshot := run
				snapshot.Observations = append([]result.Observation(nil), run.Observations...)
				if err := observe(snapshot); err != nil {
					return run, fmt.Errorf("record completed observation: %w", err)
				}
			}
			if err := ctx.Err(); err != nil {
				return run, fmt.Errorf("after %s repetition %d: %w", c.Name, repetition, err)
			}
		}
	}
	return run, nil
}

func measureContext(ctx context.Context, backend Backend, prepared Prepared, benchtime string) (Sample, error) {
	if contextual, ok := backend.(ContextBackend); ok {
		return contextual.MeasureContext(ctx, prepared, benchtime)
	}
	return backend.Measure(prepared, benchtime)
}
