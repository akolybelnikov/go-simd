package result

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
)

const LegacyBackendID = "go-test-text/v1"
const LegacyHarnessID = "legacy-report/v1"
const LegacyPatternHarnessID = "legacy-report-patterns/v1"
const LegacyReductionHarnessID = "legacy-report-reduction/v1"
const LegacyDotHarnessID = "legacy-report-dot/v1"
const LegacyFMAHarnessID = "legacy-report-affine-fma/v1"
const LegacyDotFMAHarnessID = "legacy-report-dot-fma/v1"
const LegacyPredicateHarnessID = "legacy-report-predicate-export/v1"
const LegacyPackedPredicateHarnessID = "legacy-report-packed-predicate/v1"
const LegacyCopyHarnessID = "legacy-report-copy/v1"
const LegacyClampHarnessID = "legacy-report-clamp/v1"
const LegacyXorBitsHarnessID = "legacy-report-xor-bits/v1"
const LegacyBitRoundTripHarnessID = "legacy-report-bit-roundtrip/v1"
const LegacySaturatingAdd8HarnessID = "legacy-report-saturating-add8/v1"
const LegacyAddition64HarnessID = "legacy-report-addition64/v1"
const LegacyMulAddHarnessID = "legacy-report-muladd-fma/v1"

type LegacyOptions struct {
	InputOptions *input.Options
	RunID        string
	Metadata     string
	// Set only for a fresh run of this checkout's known generated fixtures.
	// Historical text alone cannot establish its input fingerprint.
	GeneratedInputs bool
}

// ReadLegacy is the transitional importer for Go's human-readable benchmark
// format. It retains individual samples, available counters, and source names.
func ReadLegacy(reader io.Reader, options LegacyOptions) (Run, error) {
	if options.RunID == "" {
		return Run{}, fmt.Errorf("a run ID is required")
	}
	run := Run{ID: options.RunID, Environment: Environment{ID: options.RunID + "/environment"},
		BackendID: LegacyBackendID, HarnessID: LegacyHarnessID,
		SourceFormat: "go-benchmark-text", Metadata: options.Metadata}
	counts := map[MeasurementKey]int{}
	fingerprints := map[string]string{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.HasPrefix(line, "input configuration: ") {
			var decoded input.Options
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "input configuration: ")), &decoded); err != nil {
				return Run{}, fmt.Errorf("invalid input configuration: %w", err)
			}
			decoded.SumInput = decoded.SumFixture()
			decoded.DotInput = decoded.DotFixture()
			decoded.FMAInput = decoded.FMAFixture()
			var expected input.Options
			if options.InputOptions != nil {
				expected = *options.InputOptions
				expected.SumInput = expected.SumFixture()
				expected.DotInput = expected.DotFixture()
				expected.FMAInput = expected.FMAFixture()
			}
			if options.InputOptions != nil && expected != decoded {
				return Run{}, fmt.Errorf("conflicting input configuration")
			}
			options.InputOptions = &decoded
		}
		for prefix, destination := range map[string]*string{
			"goos:": &run.Environment.GOOS, "goarch:": &run.Environment.GOARCH, "cpu:": &run.Environment.CPU,
		} {
			if strings.HasPrefix(line, prefix) {
				value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if *destination != "" && value != *destination {
					return Run{}, fmt.Errorf("line %d: conflicting %s environment metadata", lineNumber, prefix)
				}
				*destination = value
			}
		}
		fields := strings.Fields(line)
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		parts := strings.Split(fields[0], "/")
		if len(parts) != 3 || !strings.HasPrefix(parts[1], "n=") {
			continue
		}
		size, err := strconv.Atoi(strings.TrimPrefix(parts[1], "n="))
		if err != nil || size < 0 {
			return Run{}, fmt.Errorf("line %d: invalid input size", lineNumber)
		}
		label, procs := benchmarkVariant(parts[2])
		metrics := map[string]Metric{}
		for i := 2; i+1 < len(fields); i += 2 {
			unit := fields[i+1]
			if strings.HasPrefix(unit, "output-sha256-") {
				value, err := strconv.ParseFloat(fields[i], 64)
				if err != nil || !finite(value) || value < 0 {
					return Run{}, fmt.Errorf("line %d: invalid output digest metric", lineNumber)
				}
				if _, exists := metrics[unit]; exists {
					return Run{}, fmt.Errorf("duplicate output digest metric")
				}
				metrics[unit] = Metric{value, true}
			}
			switch unit {
			case "ns/op", "B/op", "allocs/op", "MB/s", "simd-lanes", "result-bits", "fma-path", "fma-vector", "last-result-bits":
				value, err := strconv.ParseFloat(fields[i], 64)
				if err != nil || !finite(value) || value < 0 {
					return Run{}, fmt.Errorf("line %d: invalid %s measurement", lineNumber, unit)
				}
				metrics[unit] = Metric{value, true}
			}
		}
		if !metrics["ns/op"].Available {
			continue
		}
		iterations, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || iterations == 0 {
			return Run{}, fmt.Errorf("line %d: invalid iteration count", lineNumber)
		}
		c := legacyCase(parts[0], label, size, procs)
		harness := LegacyHarnessID
		op, knownOp := catalog.ByBenchmark(parts[0])
		var scalarBits *uint32
		if knownOp && op.ScalarResult() {
			harness = LegacyReductionHarnessID
			if op.ID == catalog.Dot {
				harness = LegacyDotHarnessID
			}
			if op.ID == catalog.DotFMA {
				harness = LegacyDotFMAHarnessID
			}
			lanes, bits := metrics["simd-lanes"], metrics["result-bits"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("%s measurement requires an integer simd-lanes metric", op.Group)
			}
			if !bits.Available || bits.Value > math.MaxUint32 || math.Trunc(bits.Value) != bits.Value {
				return Run{}, fmt.Errorf("%s measurement requires a uint32 result-bits metric", op.Group)
			}
			fold, err := ReductionFold(op.ID, c.StrategyID, int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Reduction = fold
			value := uint32(bits.Value)
			scalarBits = &value
		}
		if knownOp && op.Parameterized() {
			harness = LegacyPatternHarnessID
			c.Parameters = "unknown"
			if options.InputOptions != nil {
				c.Parameters = options.InputOptions.Identity(op)
			}
		}
		var arrayBits *uint32
		byteDigest := ""
		arrayDigest := ""
		if knownOp && (op.ID == catalog.BitRoundTrip || op.ID == catalog.XorBits) {
			harness = LegacyBitRoundTripHarnessID
			if op.ID == catalog.XorBits {
				harness = LegacyXorBitsHarnessID
			}
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("bit representation requires actual integral Float32 lanes")
			}
			identity, err := NewVectorIdentity(int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Vector = identity
			arrayDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, err
			}
		}
		if knownOp && op.ID == catalog.SaturatingAdd8 {
			harness = LegacySaturatingAdd8HarnessID
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("saturatingadd8 requires actual integral Uint8 lanes")
			}
			identity, err := NewVectorIdentity(int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Vector = identity
			byteDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, err
			}
		}
		if knownOp && op.ID == catalog.Add64 {
			harness = LegacyAddition64HarnessID
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("addition64 requires integral actual float64 lanes")
			}
			identity, err := NewVectorIdentity(int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Vector = identity
			arrayDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, err
			}
		}
		if knownOp && op.ID == catalog.MulAdd {
			harness = LegacyMulAddHarnessID
			vector, path, lanes := metrics["fma-vector"], metrics["fma-path"], metrics["simd-lanes"]
			if !vector.Available || !path.Available || vector.Value != 0 && vector.Value != 1 || path.Value != 0 && path.Value != 1 || !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("muladd requires exact gate/path/width metrics")
			}
			vectorPath := fma32.SoftwarePath
			if vector.Value == 1 {
				vectorPath = fma32.VectorPath
			}
			identity, err := NewFMAIdentity(c.StrategyID, vectorPath, int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			expectedPath := float64(0)
			if identity.PathID == fma32.VectorPath {
				expectedPath = 1
			}
			if path.Value != expectedPath {
				return Run{}, fmt.Errorf("conflicting muladd strategy/path metrics")
			}
			c.FMA = identity
			arrayDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, fmt.Errorf("muladd %w", err)
			}
		}
		if knownOp && op.ID == catalog.Clamp {
			harness = LegacyClampHarnessID
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("clamp requires an integer simd-lanes metric")
			}
			identity, err := NewClampIdentity(int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Clamp = identity
			arrayDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, fmt.Errorf("clamp %w", err)
			}
		}
		if knownOp && op.ID == catalog.Copy {
			harness = LegacyCopyHarnessID
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("copy requires an integer simd-lanes metric")
			}
			identity, err := NewCopyIdentity(int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			c.Copy = identity
			arrayDigest, err = outputDigest(metrics)
			if err != nil {
				return Run{}, fmt.Errorf("copy %w", err)
			}
		}
		if knownOp && op.ByteResult() && op.TypeID == catalog.Float32 {
			harness = LegacyPredicateHarnessID
			lanes := metrics["simd-lanes"]
			if !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value {
				return Run{}, fmt.Errorf("predicate requires an integer simd-lanes metric")
			}
			identity, err := NewPredicateIdentity(int(lanes.Value))
			if op.ID == catalog.PredicateGTPacked {
				harness = LegacyPackedPredicateHarnessID
				identity, err = NewPackedPredicateIdentity(int(lanes.Value))
			}
			if err != nil {
				return Run{}, err
			}
			c.Predicate = identity
			byteDigest, err = predicateDigest(metrics)
			if err != nil {
				return Run{}, err
			}
		}
		if knownOp && op.ID == catalog.DotFMA {
			vector, path := metrics["fma-vector"], metrics["fma-path"]
			if !vector.Available || vector.Value != 0 && vector.Value != 1 || !path.Available || path.Value > 4 || math.Trunc(path.Value) != path.Value {
				return Run{}, fmt.Errorf("fused dot requires exact fma-vector/fma-path metrics")
			}
			vectorPath := fma32.SoftwarePath
			if vector.Value == 1 {
				vectorPath = fma32.VectorPath
			}
			identity, err := NewDotFMAIdentity(c.StrategyID, vectorPath, c.Reduction.NativeLanes.Value, size)
			if err != nil {
				return Run{}, err
			}
			code, err := DotFMAPathCode(identity.PathID)
			if err != nil {
				return Run{}, err
			}
			if path.Value != float64(code) {
				return Run{}, fmt.Errorf("conflicting fused-dot case/path metrics")
			}
			c.FMA = identity
			if len(run.Observations) == 0 {
				run.Metadata += "Fused-dot child vector gate: " + vectorPath + "\n"
			}
		}
		if knownOp && op.ID == catalog.AffineFMA {
			harness = LegacyFMAHarnessID
			vector, path, lanes, bits := metrics["fma-vector"], metrics["fma-path"], metrics["simd-lanes"], metrics["last-result-bits"]
			if !vector.Available || !path.Available || vector.Value != 0 && vector.Value != 1 || path.Value != 0 && path.Value != 1 || !lanes.Available || lanes.Value < 1 || lanes.Value >= float64(int(^uint(0)>>1)) || math.Trunc(lanes.Value) != lanes.Value || !bits.Available || bits.Value > math.MaxUint32 || math.Trunc(bits.Value) != bits.Value {
				return Run{}, fmt.Errorf("fused affine requires exact path, width and last-result-bits metrics")
			}
			vectorPath := fma32.SoftwarePath
			if vector.Value == 1 {
				vectorPath = fma32.VectorPath
			}
			identity, err := NewFMAIdentity(c.StrategyID, vectorPath, int(lanes.Value))
			if err != nil {
				return Run{}, err
			}
			expectedPath := float64(0)
			if identity.PathID == fma32.VectorPath {
				expectedPath = 1
			}
			if path.Value != expectedPath {
				return Run{}, fmt.Errorf("conflicting FMA32 strategy/path metrics")
			}
			c.FMA = identity
			value := uint32(bits.Value)
			arrayBits = &value
			if len(run.Observations) == 0 {
				run.Metadata += "FMA32 child vector path: " + vectorPath + "\n"
			}
		}
		if options.GeneratedInputs && knownOp && op.ID == catalog.SaturatingAdd8 {
			c.Input = InputIdentity{input.SaturatingAdd8GeneratorID, input.SaturatingAdd8Fingerprint(size)}
		}
		if options.GeneratedInputs && knownOp && op.ID == catalog.Add64 {
			c.Input = InputIdentity{input.Addition64GeneratorID, input.Addition64Fingerprint(size)}
		}
		if options.GeneratedInputs && c.TypeID == "float32" {
			cacheKey := fmt.Sprintf("%s/%d", c.OperationID, size)
			fingerprint, ok := fingerprints[cacheKey]
			if !ok {
				if op.ID == catalog.XorBits {
					fingerprint = input.FingerprintForOperation(op, size, input.DefaultOptions())
				} else if op.ID == catalog.MulAdd {
					if options.InputOptions != nil {
						fingerprint = input.FingerprintForOperation(op, size, *options.InputOptions)
					}
				} else if op.ID == catalog.Copy || op.ID == catalog.BitRoundTrip {
					fingerprint = input.FingerprintForOperation(op, size, input.DefaultOptions())
				} else if !op.Parameterized() && !op.ScalarResult() {
					fingerprint = input.GeneratedFloat32Fingerprint(size)
				} else if options.InputOptions != nil {
					fingerprint = input.FingerprintForOperation(op, size, *options.InputOptions)
				}
				fingerprints[cacheKey] = fingerprint
			}
			if op.ID == catalog.XorBits {
				c.Input = InputIdentity{input.XorBitsGeneratorID, fingerprint}
			} else if op.ID == catalog.MulAdd {
				if options.InputOptions != nil {
					c.Input = InputIdentity{options.InputOptions.Generator(op), fingerprint}
				}
			} else if op.ID == catalog.Copy || op.ID == catalog.BitRoundTrip {
				c.Input = InputIdentity{input.DefaultOptions().Generator(op), fingerprint}
			} else if !op.Parameterized() && !op.ScalarResult() {
				c.Input = InputIdentity{input.GeneratedFloat32ID, fingerprint}
			} else if options.InputOptions != nil {
				c.Input = InputIdentity{options.InputOptions.Generator(op), fingerprint}
			}
		}
		if len(run.Observations) == 0 {
			run.HarnessID = harness
		} else if run.HarnessID != harness {
			run.HarnessID = "mixed"
		}
		key := MeasurementKey{run.ID, run.Environment.ID, run.BackendID, harness, procs, c}
		counts[key]++
		run.Observations = append(run.Observations, Observation{
			Key: key, Label: label, Repetition: counts[key], Iterations: iterations,
			NanosecondsPerOp: metrics["ns/op"].Value,
			BytesPerOp:       metrics["B/op"], AllocsPerOp: metrics["allocs/op"], MegabytesPerSec: metrics["MB/s"],
			SourceBenchmark:     fields[0],
			ScalarResultBits:    scalarBits,
			ArrayResultLastBits: arrayBits,
			ByteResultSHA256:    byteDigest,
			ArrayResultSHA256:   arrayDigest,
		})
	}
	if err := scanner.Err(); err != nil {
		return Run{}, err
	}
	if len(run.Observations) == 0 {
		return Run{}, fmt.Errorf("no benchmark measurements found")
	}
	return run, nil
}

func benchmarkVariant(name string) (string, OptionalInt) {
	if i := strings.LastIndex(name, "-"); i >= 0 {
		if n, err := strconv.Atoi(name[i+1:]); err == nil && n > 0 {
			return name[:i], OptionalInt{n, true}
		}
	}
	return name, OptionalInt{}
}

// Known aliases/policies come from the same metadata as executable registration.
// Unknown names retain their historical namespace without executable fallback.
func legacyCase(benchmark, label string, size int, procs OptionalInt) Case {
	c := Case{OperationID: "legacy:" + benchmark, TypeID: "unknown", StrategyID: "legacy:" + label,
		ArithmeticID: "unknown", Elements: size, ExecutionMode: catalog.KernelOnly}
	if op, ok := catalog.ByBenchmark(benchmark); ok {
		c.OperationID, c.TypeID, c.ArithmeticID, c.ExecutionMode = op.ID, op.TypeID, op.ArithmeticID, op.ExecutionMode
	}
	if variant, ok := catalog.LegacyVariant(label); ok {
		c.StrategyID = variant.ID
		if c.TypeID != "unknown" {
			knownProcs := 0
			if procs.Known {
				knownProcs = procs.Value
			}
			count, known := variant.WorkerCount(size, knownProcs)
			c.Workers = OptionalInt{count, known}
		}
	}
	return c
}
