package patterns_test

import (
	"math"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func TestFusedAffineRoundingAndPath(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	u := float32(0x1p-23)
	if simd.Emulated() && patterns.FMAPath() != fma32.SoftwarePath {
		t.Fatal("emulation mislabeled as fused vectors")
	}
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
		for _, p := range [][3]float32{
			{1 + u, 1 - u, -1}, {1 + u, 1 - u, 16777218},
			{float32(math.Ldexp(float64(1+u), -75)), float32(math.Ldexp(float64(1-u), -75)), math.Float32frombits(0x007fffff)},
			{float32(math.Ldexp(float64(1+u), 51)), float32(math.Ldexp(float64(1-u), 52)), math.MaxFloat32},
		} {
			x := make([]float32, n)
			for i := range x {
				x[i] = p[1]
			}
			want := math.Float32bits(fma32.FMA(p[0], p[1], p[2]))
			for _, kernel := range []func([]float32, []float32, float32, float32) error{patterns.AffineFMAScalar, patterns.AffineFMASIMD} {
				out := make([]float32, n)
				if err := kernel(x, out, p[0], p[2]); err != nil {
					t.Fatal(err)
				}
				for _, got := range out {
					if math.Float32bits(got) != want {
						t.Fatalf("fused rounding differs n=%d: %08x != %08x", n, math.Float32bits(got), want)
					}
				}
				if got := testing.AllocsPerRun(10, func() {
					if err := kernel(x, out, p[0], p[2]); err != nil {
						panic(err)
					}
				}); got != 0 {
					t.Fatal("FMA kernel allocation", got)
				}
			}
		}
	}
}

func TestFMAKernelRandomBitsAgainstVerifiedReference(t *testing.T) {
	var vector simd.Float32s
	n := 2*vector.Len() + 1
	rng := rand.New(rand.NewPCG(317, 902))
	x, out := make([]float32, n), make([]float32, n)
	for trial := 0; trial < 512; trial++ {
		alpha, beta := math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32())
		for i := range x {
			x[i] = math.Float32frombits(rng.Uint32())
		}
		for _, kernel := range []func([]float32, []float32, float32, float32) error{patterns.AffineFMAScalar, patterns.AffineFMASIMD} {
			if err := kernel(x, out, alpha, beta); err != nil {
				t.Fatal(err)
			}
			for i, v := range x {
				want := fma32.FMA(alpha, v, beta)
				if math.IsNaN(float64(want)) {
					if !math.IsNaN(float64(out[i])) {
						t.Fatal("missing NaN")
					}
					continue
				}
				if math.Float32bits(out[i]) != math.Float32bits(want) {
					t.Fatalf("random FMA mismatch trial=%d lane=%d: got %08x want %08x", trial, i, math.Float32bits(out[i]), math.Float32bits(want))
				}
			}
		}
	}
}
