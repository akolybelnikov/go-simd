package result

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
)

const LegacyBackendID = "go-test-text/v1"
const LegacyHarnessID = "legacy-report/v1"

type LegacyOptions struct {
	RunID    string
	Metadata string
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
	fingerprints := map[int]string{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
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
			switch unit {
			case "ns/op", "B/op", "allocs/op", "MB/s":
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
		if options.GeneratedInputs && c.TypeID == "float32" {
			fingerprint, ok := fingerprints[size]
			if !ok {
				fingerprint = input.GeneratedFloat32Fingerprint(size)
				fingerprints[size] = fingerprint
			}
			c.Input = InputIdentity{input.GeneratedFloat32ID, fingerprint}
		}
		key := MeasurementKey{run.ID, run.Environment.ID, run.BackendID, run.HarnessID, procs, c}
		counts[key]++
		run.Observations = append(run.Observations, Observation{
			Key: key, Label: label, Repetition: counts[key], Iterations: iterations,
			NanosecondsPerOp: metrics["ns/op"].Value,
			BytesPerOp:       metrics["B/op"], AllocsPerOp: metrics["allocs/op"], MegabytesPerSec: metrics["MB/s"],
			SourceBenchmark: fields[0],
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
		if c.TypeID == catalog.Float32 {
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
