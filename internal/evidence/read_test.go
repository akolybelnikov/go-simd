package evidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func validCapture() Capture {
	duration := int64(50)
	key := result.MeasurementKey{
		RunID: "run", EnvironmentID: "env", BackendID: "backend", HarnessID: "harness",
		Case: result.Case{OperationID: catalog.Add, TypeID: catalog.Float32, StrategyID: catalog.Scalar, Elements: 8, Input: result.InputIdentity{GeneratorID: "generated-float32-v1"}},
	}
	return Capture{
		Schema: Schema, Status: "complete", Group: "addition",
		Config:  Config{Count: 1, Benchtime: "1x", Bench: "."},
		Machine: Machine{GoVersion: "go-test", OS: "windows", Arch: "amd64", CPU: "test", Bits: 256, Lanes: 8, Procs: 4},
		Run:     result.Run{ID: "run", Environment: result.Environment{ID: "env", GOOS: "windows", GOARCH: "amd64", CPU: "test"}, BackendID: "backend", HarnessID: "harness", SourceFormat: Schema, Observations: []result.Observation{{Key: key, Label: "Scalar", Repetition: 1, Iterations: 5, DurationNS: &duration, NanosecondsPerOp: 10, SourceBenchmark: "BenchmarkAddition/n=8/Scalar-4"}}},
	}
}

func encodeCapture(t *testing.T, capture Capture) []byte {
	t.Helper()
	data, err := json.Marshal(capture)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func readCode(err error) string {
	var target *ReadError
	if errors.As(err, &target) {
		return target.Code
	}
	return ""
}

func TestReadCaptureRoundTripAndBounds(t *testing.T) {
	data := encodeCapture(t, validCapture())
	got, err := ReadCapture(bytes.NewReader(data), DefaultReadLimits())
	if err != nil || len(got.Run.Observations) != 1 {
		t.Fatalf("capture=%+v err=%v", got, err)
	}
	limits := DefaultReadLimits()
	limits.MaxBytes = int64(len(data) - 1)
	if _, err := ReadCapture(bytes.NewReader(data), limits); readCode(err) != "capture_too_large" {
		t.Fatalf("size error=%v", err)
	}
	limits = DefaultReadLimits()
	limits.MaxObservations = 1
	capture := validCapture()
	capture.Run.Observations = append(capture.Run.Observations, capture.Run.Observations[0])
	if _, err := ReadCapture(bytes.NewReader(encodeCapture(t, capture)), limits); readCode(err) != "too_many_observations" {
		t.Fatalf("observation limit error=%v", err)
	}
}

func TestReadCaptureRejectsMalformedFutureAndIncomplete(t *testing.T) {
	if _, err := ReadCapture(strings.NewReader(`{"schema":"future"}`), DefaultReadLimits()); readCode(err) != "unsupported_schema" {
		t.Fatalf("future schema error=%v", err)
	}
	if _, err := ReadCapture(strings.NewReader(`{"schema":"go-simd.measurements/v1","unknown":true}`), DefaultReadLimits()); readCode(err) != "invalid_json" {
		t.Fatalf("unknown field error=%v", err)
	}
	capture := validCapture()
	capture.Run.Observations = nil
	if _, err := ReadCapture(bytes.NewReader(encodeCapture(t, capture)), DefaultReadLimits()); readCode(err) != "incomplete_capture" {
		t.Fatalf("incomplete error=%v", err)
	}
	capture = validCapture()
	capture.Run.Observations[0].Key.RunID = "other"
	if _, err := ReadCapture(bytes.NewReader(encodeCapture(t, capture)), DefaultReadLimits()); readCode(err) != "identity_mismatch" {
		t.Fatalf("identity error=%v", err)
	}
}

func TestReadCaptureAcceptsFailedEnvelopeWithoutObservations(t *testing.T) {
	capture := Capture{Schema: Schema, Status: "failed", Error: "plan rejected", Group: "addition"}
	if _, err := ReadCapture(bytes.NewReader(encodeCapture(t, capture)), DefaultReadLimits()); err != nil {
		t.Fatal(err)
	}
}
