package engine

import (
	"errors"
	"flag"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/result"
)

func TestConfigurationValidation(t *testing.T) {
	for _, value := range []string{"1x", "7x", "1ns", "1.5ms", "100ms"} {
		if err := (Config{1, value, "."}).Validate(); err != nil {
			t.Errorf("%s: %v", value, err)
		}
	}
	for _, value := range []string{"", "0x", "-1x", "1.5x", "999999999999999999999x", "0s", "-1ms", "invalid", "0.1ns"} {
		if err := (Config{1, value, "."}).Validate(); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if err := (Config{0, "1x", "."}).Validate(); err == nil {
		t.Fatal("zero count accepted")
	}
	if err := (Config{1, "1x", "["}).Validate(); err == nil {
		t.Fatal("invalid regex accepted")
	}
}

func TestGoSelectionGrammar(t *testing.T) {
	for _, tt := range []struct {
		expression, name string
		want             bool
	}{
		{".", "BenchmarkAddition/n=8/Scalar", true},
		{"", "BenchmarkAddition/n=8/Scalar", false},
		{"Addition/n=8/Scalar", "BenchmarkAddition/n=8/Scalar", true},
		{"^BenchmarkAddition$/^n=(8|16)$/^(Scalar|SIMDTail)$", "BenchmarkAddition/n=16/SIMDTail", true},
		{"^BenchmarkAddition$/^n=(8|16)$/^(Scalar|SIMDTail)$", "BenchmarkAddition/n=128/Scalar", false},
		{"^BenchmarkAddition$/^n=[8/9]$/^Scalar$", "BenchmarkAddition/n=8/Scalar", true},
		{`^BenchmarkAddition$/^n=8$/Scalar\/x`, "BenchmarkAddition/n=8/Scalar", false},
		{"^BenchmarkAddition$/^n=8$/^Scalar$|^BenchmarkComputation$/^n=16$/^SIMD$", "BenchmarkComputation/n=16/SIMD", true},
		{"^BenchmarkAddition$/^n=8$/^Scalar$/extra", "BenchmarkAddition/n=8/Scalar", false},
		{"Addition /n=8", "BenchmarkAddition/n=8/Scalar", false},
	} {
		match, err := compileSelection(tt.expression)
		if err != nil {
			t.Fatal(err)
		}
		if got := match(tt.name); got != tt.want {
			t.Errorf("%q on %q = %t", tt.expression, tt.name, got)
		}
	}
}

type fakeBackend struct{ calls, fail int }

func (b *fakeBackend) Measure(p Prepared, _ string) (Sample, error) {
	b.calls++
	if b.calls == b.fail {
		return Sample{}, errors.New("injected kernel failure")
	}
	if err := p.Run(); err != nil {
		return Sample{}, err
	}
	return Sample{Iterations: 5, Duration: 2 * time.Nanosecond, MemAllocs: 2, MemBytes: 7}, nil
}

func TestDirectObservationsAndPartialFailure(t *testing.T) {
	plan, err := Prepare("addition", Config{2, "5x", "^BenchmarkAddition$/^n=8$/^(Scalar|SIMDParallel)$"}, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Cases) != 2 {
		t.Fatalf("selected cases=%d", len(plan.Cases))
	}
	run := result.Run{ID: "run", Environment: result.Environment{ID: "env"}, BackendID: BackendID, HarnessID: HarnessID}
	complete, err := Execute(plan, run, &fakeBackend{})
	if err != nil || len(complete.Observations) != 4 {
		t.Fatalf("%v, %+v", err, complete)
	}
	for i, o := range complete.Observations {
		if o.DurationNS == nil || *o.DurationNS != 2 || o.Iterations != 5 || o.NanosecondsPerOp != 0.4 || o.BytesPerOp.Value != 1.4 || o.AllocsPerOp.Value != 0.4 || o.Repetition != i%2+1 || o.Key.Case.Input.SHA256 == "" || o.Key.Case.Input.GeneratorID != "generated-float32-v1" {
			t.Fatalf("precision/identity lost: %+v", o)
		}
		wantWorkers := 1
		if i >= 2 {
			wantWorkers = min(8, runtime.GOMAXPROCS(0))
		}
		if !o.Key.Case.Workers.Known || o.Key.Case.Workers.Value != wantWorkers || o.Key.HarnessID != HarnessID || o.MegabytesPerSec.Value != 240000 {
			t.Fatalf("accounting lost: %+v", o)
		}
	}
	partial, err := Execute(plan, run, &fakeBackend{fail: 2})
	if err == nil || !strings.Contains(err.Error(), "injected kernel failure") || len(partial.Observations) != 1 {
		t.Fatalf("partial=%+v error=%v", partial, err)
	}
	if _, err := Execute(Plan{}, run, &fakeBackend{}); err == nil {
		t.Fatal("empty selection accepted")
	}
}

func TestTestingAdapterModesAndFlagRestoration(t *testing.T) {
	previous := flag.Lookup("test.benchtime").Value.String()
	for _, mode := range []string{"1x", "7x", "1ms"} {
		out := []float32{0}
		sample, err := (Testing{}).Measure(Prepared{Run: func() error { out[0]++; return nil }, Output: out, Bytes: 4}, mode)
		if err != nil || sample.Iterations < 1 || sample.Duration < 0 || out[0] < 1 {
			t.Fatalf("%s: %+v %v", mode, sample, err)
		}
		if mode == "7x" && sample.Iterations != 7 {
			t.Fatalf("iteration mode ignored: %+v", sample)
		}
		if flag.Lookup("test.benchtime").Value.String() != previous {
			t.Fatal("testing flag was not restored")
		}
	}
	want := errors.New("unexpected kernel failure")
	_, err := (Testing{}).Measure(Prepared{Run: func() error { return want }, Output: []float32{0}}, "1x")
	if !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
	if flag.Lookup("test.benchtime").Value.String() != previous {
		t.Fatal("failure did not restore timing")
	}
}

func TestTestingAdapterSerializesSettings(t *testing.T) {
	var wg sync.WaitGroup
	for _, mode := range []struct {
		s string
		n int
	}{{"2x", 2}, {"3x", 3}} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out := []float32{0}
			sample, err := (Testing{}).Measure(Prepared{Run: func() error { out[0]++; return nil }, Output: out}, mode.s)
			if err != nil || sample.Iterations != mode.n {
				t.Errorf("%s: %+v %v", mode.s, sample, err)
			}
		}()
	}
	wg.Wait()
}
