package result

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestReadLegacyRetainsObservationIdentity(t *testing.T) {
	raw := "goos: linux\ngoarch: arm64\ncpu: measured CPU\n" +
		"BenchmarkAddition/n=9/SIMDParallel-4 123 1.234567 ns/op 12.5 MB/s 128 B/op 2 allocs/op\n" +
		"BenchmarkAddition/n=9/SIMDParallel-4 456 2.345678 ns/op 0 B/op 0 allocs/op\n" +
		"BenchmarkAddition/n=9/SIMDParallel-2 789 3.456789 ns/op\n" +
		"BenchmarkComputation/n=9/Scalar-4 123 4.567891 ns/op\n"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "run-1", Metadata: "parent context"})
	if err != nil {
		t.Fatal(err)
	}
	if run.Environment != (Environment{"run-1/environment", "linux", "arm64", "measured CPU"}) || run.Metadata != "parent context" {
		t.Fatalf("environment/provenance lost: %+v", run)
	}
	if run.SourceFormat != "go-benchmark-text" || run.BackendID != LegacyBackendID || run.HarnessID != LegacyHarnessID {
		t.Fatalf("source identity lost: %+v", run)
	}
	if len(run.Observations) != 4 {
		t.Fatalf("observations = %d", len(run.Observations))
	}
	first, second, third, scalar := run.Observations[0], run.Observations[1], run.Observations[2], run.Observations[3]
	if first.Repetition != 1 || second.Repetition != 2 || third.Repetition != 1 || scalar.Repetition != 1 {
		t.Fatal("repetition IDs must be per complete case/context")
	}
	if first.Iterations != 123 || second.Iterations != 456 || first.NanosecondsPerOp != 1.234567 || first.DurationNS != nil {
		t.Fatal("iterations/precision changed or exact duration invented")
	}
	if first.BytesPerOp != (Metric{128, true}) || first.AllocsPerOp != (Metric{2, true}) || first.MegabytesPerSec != (Metric{12.5, true}) || second.BytesPerOp != (Metric{0, true}) || second.MegabytesPerSec.Available {
		t.Fatal("optional metrics lost or missing confused with zero")
	}
	if first.Key.RunID != run.ID || first.Key.EnvironmentID != run.Environment.ID || first.Key.Case.OperationID != "add" || first.Key.Case.StrategyID != "simd-parallel-scalar-tail" || first.Key.Case.Workers != (OptionalInt{4, true}) || third.Key.Case.Workers != (OptionalInt{2, true}) {
		t.Fatal("typed case/run/environment identity lost")
	}
	if first.Key.Case.Input != (InputIdentity{}) || scalar.Key.Case.ArithmeticID != "go-float32-quadratic-v1" || scalar.Key.Case.Workers != (OptionalInt{1, true}) {
		t.Fatal("legacy input provenance or arithmetic/worker identity incorrect")
	}
}

func TestReadLegacyFreshGeneratedInput(t *testing.T) {
	raw := "BenchmarkAddition/n=8/Scalar-16 1 1 ns/op\nBenchmarkAddition/n=8/SIMDPartial-16 1 2 ns/op\n"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "fresh", GeneratedInputs: true})
	if err != nil {
		t.Fatal(err)
	}
	want := InputIdentity{input.GeneratedFloat32ID, input.GeneratedFloat32Fingerprint(8)}
	for _, o := range run.Observations {
		if o.Key.Case.Input != want {
			t.Fatalf("input identity = %+v, want %+v", o.Key.Case.Input, want)
		}
	}
}

func TestLegacySuffixAndUnknownNames(t *testing.T) {
	raw := "BenchmarkOther/n=2/custom-hyphen-name-8 1 1 ns/op\nBenchmarkOther/n=2/another-name 1 2 ns/op\n"
	run, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "unknown"})
	if err != nil {
		t.Fatal(err)
	}
	if run.Observations[0].Label != "custom-hyphen-name" || run.Observations[1].Label != "another-name" {
		t.Fatal("only numeric GOMAXPROCS suffixes should be stripped")
	}
	if run.Observations[0].Key.Case.OperationID != "legacy:BenchmarkOther" || run.Observations[0].Key.Case.Workers.Known || run.Observations[1].Key.GOMAXPROCS.Known {
		t.Fatal("unknown metadata was invented")
	}
}

func TestReadLegacyRejectsMalformedMeasurements(t *testing.T) {
	for _, raw := range []string{
		"PASS",
		"BenchmarkAddition/n=bad/Scalar-8 1 2 ns/op",
		"BenchmarkAddition/n=-1/Scalar-8 1 2 ns/op",
		"BenchmarkAddition/n=1/Scalar-8 0 2 ns/op",
		"BenchmarkAddition/n=1/Scalar-8 bad 2 ns/op",
		"BenchmarkAddition/n=1/Scalar-8 1 NaN ns/op",
		"BenchmarkAddition/n=1/Scalar-8 1 +Inf ns/op",
		"BenchmarkAddition/n=1/Scalar-8 1 -1 ns/op",
		"BenchmarkAddition/n=1/Scalar-8 1 2 ns/op NaN B/op",
		"cpu: first\nBenchmarkAddition/n=1/Scalar-8 1 2 ns/op\ncpu: second\nBenchmarkAddition/n=1/Scalar-8 1 3 ns/op",
	} {
		if _, err := ReadLegacy(strings.NewReader(raw), LegacyOptions{RunID: "invalid"}); err == nil {
			t.Errorf("accepted invalid raw text: %s", raw)
		}
	}
	if _, err := ReadLegacy(strings.NewReader(""), LegacyOptions{}); err == nil {
		t.Fatal("missing run identity accepted")
	}
	want := errors.New("source read failed")
	if _, err := ReadLegacy(failingReader{want}, LegacyOptions{RunID: "broken"}); !errors.Is(err, want) {
		t.Fatalf("reader error = %v", err)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

var _ io.Reader = failingReader{}
