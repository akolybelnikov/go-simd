package evidence

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/result"
)

type ReadLimits struct {
	MaxBytes        int64
	MaxObservations int
}

func DefaultReadLimits() ReadLimits {
	return ReadLimits{MaxBytes: 8 << 20, MaxObservations: 10_000}
}

type ReadError struct {
	Code    string
	Message string
}

func (e *ReadError) Error() string { return e.Message }

func readError(code, format string, args ...any) error {
	return &ReadError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// ReadCapture decodes one bounded v1 capture, rejects unknown fields/future
// schemas, and validates observation identity before analysis consumes it.
func ReadCapture(reader io.Reader, limits ReadLimits) (Capture, error) {
	if reader == nil {
		return Capture{}, readError("invalid_reader", "capture reader is required")
	}
	if limits.MaxBytes < 1 || limits.MaxObservations < 1 {
		return Capture{}, readError("invalid_limits", "capture read limits must be positive")
	}
	limited := &io.LimitedReader{R: reader, N: limits.MaxBytes + 1}
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	var capture Capture
	if err := decoder.Decode(&capture); err != nil {
		return Capture{}, readError("invalid_json", "decode measurement capture: %v", err)
	}
	if limited.N == 0 {
		return Capture{}, readError("capture_too_large", "measurement capture exceeds %d bytes", limits.MaxBytes)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Capture{}, readError("trailing_data", "measurement capture contains multiple JSON values")
		}
		return Capture{}, readError("invalid_json", "decode trailing measurement data: %v", err)
	}
	if err := validateCapture(capture, limits); err != nil {
		return Capture{}, err
	}
	return capture, nil
}

func validateCapture(capture Capture, limits ReadLimits) error {
	if capture.Schema != Schema {
		return readError("unsupported_schema", "unsupported measurement schema %q", capture.Schema)
	}
	if capture.Status != "complete" && capture.Status != "failed" {
		return readError("invalid_status", "invalid measurement status %q", capture.Status)
	}
	if capture.Status == "complete" && capture.Error != "" {
		return readError("invalid_status", "complete measurement capture contains an error")
	}
	if capture.Status == "failed" && capture.Error == "" {
		return readError("invalid_status", "failed measurement capture requires an error")
	}
	if len(capture.Run.Observations) > limits.MaxObservations {
		return readError("too_many_observations", "measurement capture has %d observations; maximum is %d", len(capture.Run.Observations), limits.MaxObservations)
	}
	if len(capture.Run.Observations) == 0 {
		if capture.Status == "complete" {
			return readError("incomplete_capture", "complete measurement capture has no observations")
		}
		return nil
	}
	if _, ok := catalog.ByGroup(capture.Group); !ok {
		return readError("unknown_group", "unknown measurement group %q", capture.Group)
	}
	if capture.Config.Count < 1 || capture.Config.Bench == "" || !validBenchtime(capture.Config.Benchtime) {
		return readError("invalid_config", "measurement count, selection and benchtime must be valid")
	}
	if capture.Machine.GoVersion == "" || capture.Machine.OS == "" || capture.Machine.Arch == "" || capture.Machine.Bits < 1 || capture.Machine.Lanes < 1 || capture.Machine.Bits != capture.Machine.Lanes*32 || capture.Machine.Procs < 1 {
		return readError("invalid_machine", "measurement machine identity is incomplete or inconsistent")
	}
	if capture.Run.ID == "" || capture.Run.Environment.ID == "" || capture.Run.BackendID == "" || capture.Run.HarnessID == "" {
		return readError("invalid_run", "measurement run identity is incomplete")
	}
	if capture.Run.SourceFormat != "" && capture.Run.SourceFormat != Schema {
		return readError("invalid_run", "measurement source format %q is inconsistent", capture.Run.SourceFormat)
	}
	return validateObservations(capture)
}

func validateObservations(capture Capture) error {
	repetitions := make(map[result.MeasurementKey]map[int]bool)
	for index, observation := range capture.Run.Observations {
		key := observation.Key
		if key.RunID != capture.Run.ID || key.EnvironmentID != capture.Run.Environment.ID || key.BackendID != capture.Run.BackendID || key.HarnessID != capture.Run.HarnessID {
			return readError("identity_mismatch", "observation %d identity does not match its run", index)
		}
		if key.Case.OperationID == "" || key.Case.TypeID == "" || key.Case.StrategyID == "" || key.Case.Elements < 1 || key.Case.Input.GeneratorID == "" {
			return readError("invalid_observation", "observation %d case identity is incomplete", index)
		}
		if observation.Repetition < 1 || observation.Repetition > capture.Config.Count || observation.Iterations < 1 || observation.DurationNS == nil || *observation.DurationNS < 0 || !finiteNonnegative(observation.NanosecondsPerOp) || observation.SourceBenchmark == "" {
			return readError("invalid_observation", "observation %d timing or repetition is invalid", index)
		}
		if !validMetric(observation.BytesPerOp) || !validMetric(observation.AllocsPerOp) || !validMetric(observation.MegabytesPerSec) {
			return readError("invalid_metric", "observation %d contains an invalid metric", index)
		}
		for _, digest := range []string{observation.ArrayResultSHA256, observation.ByteResultSHA256, key.Case.Input.SHA256} {
			if digest != "" && !validSHA256(digest) {
				return readError("invalid_digest", "observation %d contains an invalid SHA-256 digest", index)
			}
		}
		if repetitions[key] == nil {
			repetitions[key] = make(map[int]bool)
		}
		if repetitions[key][observation.Repetition] {
			return readError("duplicate_observation", "observation %d repeats measurement repetition %d", index, observation.Repetition)
		}
		repetitions[key][observation.Repetition] = true
	}
	if capture.Status == "complete" {
		for key, seen := range repetitions {
			for repetition := 1; repetition <= capture.Config.Count; repetition++ {
				if !seen[repetition] {
					return readError("incomplete_capture", "complete measurement for %s/%d/%s is missing repetition %d", key.Case.OperationID, key.Case.Elements, key.Case.StrategyID, repetition)
				}
			}
		}
	}
	return nil
}

func validBenchtime(value string) bool {
	if strings.HasSuffix(value, "x") {
		n, err := strconv.ParseUint(strings.TrimSuffix(value, "x"), 10, 64)
		return err == nil && n > 0
	}
	duration, err := time.ParseDuration(value)
	return err == nil && duration > 0
}

func validMetric(metric result.Metric) bool {
	if !metric.Available {
		return metric.Value == 0
	}
	return finiteNonnegative(metric.Value)
}

func finiteNonnegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
