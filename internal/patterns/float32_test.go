package patterns_test

import (
	"errors"
	"fmt"
	"math"
	"os"
	"simd"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func oracle(id string, x, y float32, o input.Options) float32 {
	switch id {
	case catalog.Scale:
		return o.Float("alpha") * x
	case catalog.Affine:
		return float32(o.Float("alpha")*x) + o.Float("beta")
	case catalog.ThresholdZero:
		if x > o.Float("threshold") {
			return x
		}
		return 0
	case catalog.SelectGT:
		if x > o.Float("threshold") {
			return x
		}
		return y
	case catalog.RangeZero:
		if x > o.Float("lower") && x < o.Float("upper") {
			return x
		}
		return 0
	}
	panic(id)
}

func TestPatternKernelsAgainstIndependentOracle(t *testing.T) {
	if strings.Contains(os.Getenv("GODEBUG"), "simd=0") && !simd.Emulated() {
		t.Fatal("requested emulation was not enabled")
	}
	var vector simd.Float32s
	w := vector.Len()
	values := []float32{math.Float32frombits(0x80000000), 0, 1, -1, 2, -2, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), math.Float32frombits(0x7f812345)}
	base := input.DefaultOptions()
	options := []input.Options{base}
	for _, bits := range []uint32{0, 0x80000000, 0x7f800000, 0xff800000, 0x7fc12345, math.Float32bits(-1.5)} {
		o := base
		o.Alpha = bits
		o.Beta = bits
		o.Threshold = bits
		o.Lower = bits
		options = append(options, o)
	}
	reverse := base
	reverse.Lower, reverse.Upper = base.Upper, base.Lower
	options = append(options, reverse)
	nanUpper := base
	nanUpper.Upper = 0x7fc12345
	options = append(options, nanUpper)
	for _, op := range catalog.All() {
		if !op.Parameterized() {
			continue
		}
		binding, err := kernels.Lookup(op.ID, op.TypeID)
		if err != nil {
			t.Fatal(err)
		}
		for _, implementation := range binding.Implementations {
			for k, o := range options {
				for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
					for _, alias := range []string{"disjoint", "x", "y"} {
						if alias == "y" && len(op.Inputs) != 2 {
							continue
						}
						t.Run(fmt.Sprintf("%s/%s/params=%d/n=%d/%s", op.ID, implementation.Descriptor.Label, k, n, alias), func(t *testing.T) {
							x, y, want := make([]float32, n), make([]float32, n), make([]float32, n)
							for i := range x {
								x[i] = values[i%len(values)]
								y[i] = values[(i+7)%len(values)]
								want[i] = oracle(op.ID, x[i], y[i], o)
							}
							canary := math.Float32frombits(0x4f123456)
							backing := make([]float32, n+2)
							backing[0], backing[n+1] = canary, canary
							out := backing[1 : n+1]
							if alias == "x" {
								out = x
							}
							if alias == "y" {
								out = y
							}
							if err := implementation.BindConfigured(x, y, out, o)(); err != nil {
								t.Fatal(err)
							}
							for i, v := range want {
								if (op.ID == catalog.Scale || op.ID == catalog.Affine) && math.IsNaN(float64(v)) {
									if !math.IsNaN(float64(out[i])) {
										t.Fatalf("non-NaN at %d", i)
									}
									continue
								}
								if math.Float32bits(out[i]) != math.Float32bits(v) {
									t.Fatalf("index %d: got 0x%08x, want 0x%08x", i, math.Float32bits(out[i]), math.Float32bits(v))
								}
							}
							if math.Float32bits(backing[0]) != math.Float32bits(canary) || math.Float32bits(backing[n+1]) != math.Float32bits(canary) {
								t.Fatal("partial store changed canary")
							}
						})
					}
				}
			}
			out := []float32{123}
			if err := implementation.BindConfigured([]float32{1, 2}, []float32{1, 2}, out, base)(); !errors.Is(err, patterns.ErrLengthMismatch) || out[0] != 123 {
				t.Fatal("length validation modified output")
			}
			if op.ID == catalog.SelectGT {
				if err := implementation.BindConfigured([]float32{1}, []float32{1, 2}, out, base)(); !errors.Is(err, patterns.ErrLengthMismatch) || out[0] != 123 {
					t.Fatal("second-input length check lost")
				}
			}
		}
	}
}

func TestAffineExplicitUnfusedRounding(t *testing.T) {
	x := make([]float32, 65)
	for i := range x {
		x[i] = math.Float32frombits(0x3f800001)
	}
	alpha, beta := math.Float32frombits(0x3f7ffffe), float32(-1)
	for _, kernel := range []func([]float32, []float32, float32, float32) error{patterns.AffineScalar, patterns.AffineSIMD} {
		out := make([]float32, len(x))
		if err := kernel(x, out, alpha, beta); err != nil {
			t.Fatal(err)
		}
		for _, v := range out {
			if math.Float32bits(v) != 0 {
				t.Fatalf("fused product/addition: %g", v)
			}
		}
	}
}
