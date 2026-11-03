package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/engine"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
	"math"
	"os"
	"path/filepath"
	"simd"
	"strings"
	"testing"
)

func TestClampResultsAndBackendIdentity(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, _ := catalog.ByGroup("clamp")
	for _, fixture := range []struct{ pattern, lower, upper string }{
		{"none", "-1.5", "2.5"}, {"all", "-1.5", "2.5"}, {"alternating", "-1.5", "2.5"}, {"clustered", "-1.5", "2.5"}, {"shuffled", "-1.5", "2.5"},
		{"alternating", "0", "-0"}, {"none", "-Inf", "+Inf"},
	} {
		t.Run(fixture.pattern+fixture.lower+fixture.upper, func(t *testing.T) {
			options := input.DefaultOptions()
			options.Pattern = fixture.pattern
			options.Seed = 123
			var args []string
			if err := optionsForClampTest(&options, fixture.lower, fixture.upper); err != nil {
				t.Fatal(err)
			}
			sizes := map[int]bool{1: true, w - 1: true, w: true, w + 1: true, 64: true}
			selection := fmt.Sprintf("^BenchmarkClamp$/^n=(1|%d|%d|%d|64)$", w-1, w, w+1)
			var runs []result.Run
			for _, backend := range []string{"compiled", "go-test"} {
				output := t.TempDir()
				args = []string{"-group=clamp", "-backend=" + backend, "-bench=" + selection, "-count=2", "-benchtime=3x", "-format=md", "-output", output, "-pattern=" + fixture.pattern, "-seed=123", "-lower=" + fixture.lower, "-upper=" + fixture.upper}
				stdout, stderr, err := runReportProcess(t, args...)
				if err != nil {
					t.Fatalf("%s: %v\n%s\n%s", backend, err, stdout, stderr)
				}
				files, _ := filepath.Glob(filepath.Join(output, "clamp", "*"))
				if len(files) != 2 {
					t.Fatal(files)
				}
				var run result.Run
				for _, file := range files {
					data, err := os.ReadFile(file)
					if err != nil {
						t.Fatal(err)
					}
					if strings.HasSuffix(file, ".measurements.json") {
						var c evidence.Capture
						if err := json.Unmarshal(data, &c); err != nil {
							t.Fatal(err)
						}
						run = c.Run
						if c.Status != "complete" || run.HarnessID != engine.ClampHarnessID {
							t.Fatal(c)
						}
					} else if strings.HasSuffix(file, ".raw") {
						run, err = result.ReadLegacy(bytes.NewReader(data), result.LegacyOptions{RunID: "verify", GeneratedInputs: true})
						if err != nil || run.HarnessID != result.LegacyClampHarnessID {
							t.Fatal(run, err)
						}
					}
				}
				if len(run.Observations) != len(sizes)*6 {
					t.Fatal("clamp count", len(run.Observations))
				}
				for _, o := range run.Observations {
					c := o.Key.Case
					x, _, out := input.MakeForOperation(op, c.Elements, options)
					lo, hi := options.Float("lower"), options.Float("upper")
					for i, v := range x {
						switch {
						case v < lo:
							out[i] = lo
						case v > hi:
							out[i] = hi
						default:
							out[i] = v
						}
					}
					if !sizes[c.Elements] || c.OperationID != catalog.Clamp || c.ArithmeticID != op.ArithmeticID || c.Parameters != options.Identity(op) || c.Input.GeneratorID != options.Generator(op) || c.Input.SHA256 != input.FingerprintFloat32(x, nil) || o.ArrayResultSHA256 != input.FingerprintFloat32(out, nil) || c.Clamp.PolicyID != result.ClampPolicyID || c.Clamp.NativeLanes.Value != w || c.Copy != (result.CopyIdentity{}) || o.Iterations != 3 || o.ByteResultSHA256 != "" || o.ScalarResultBits != nil {
						t.Fatal("clamp evidence lost", o)
					}
				}
				runs = append(runs, run)
			}
			for i, a := range runs[0].Observations {
				b := runs[1].Observations[i]
				if a.Key.Case != b.Key.Case || a.ArrayResultSHA256 != b.ArrayResultSHA256 {
					t.Fatal("clamp backend parity", a, b)
				}
			}
		})
	}
}

func optionsForClampTest(o *input.Options, lo, hi string) error {
	// Parse the same accepted spellings independently of the CLI flag binding.
	values := map[string]float32{"-1.5": -1.5, "2.5": 2.5, "0": 0, "-0": math.Float32frombits(0x80000000), "-Inf": float32(math.Inf(-1)), "+Inf": float32(math.Inf(1))}
	o.Lower = math.Float32bits(values[lo])
	o.Upper = math.Float32bits(values[hi])
	return o.ValidateFor(mustClampOperation())
}
func mustClampOperation() catalog.Operation { op, _ := catalog.ByGroup("clamp"); return op }

func TestClampValidationBeforeArtifacts(t *testing.T) {
	for _, args := range [][]string{{"-lower=1", "-upper=-1"}, {"-lower=NaN"}, {"-upper=NaN"}, {"-lower=-Inf", "-upper=+Inf", "-pattern=all"}} {
		output := filepath.Join(t.TempDir(), "reports")
		var stdout bytes.Buffer
		if err := run(append(append([]string{"-group=clamp"}, args...), "-output", output), &stdout); err == nil {
			t.Fatal("invalid clamp accepted", args)
		}
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatal("invalid options created artifacts")
		}
	}
}
