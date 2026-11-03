package patterns_test

import (
	"errors"
	"math"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/patterns"
)

var clampKernels = []struct {
	name string
	run  func([]float32, []float32, float32, float32) error
}{
	{"Scalar", patterns.ClampScalar}, {"SIMDPartial", patterns.ClampSIMD}, {"SIMDMinMax", patterns.ClampSIMDMinMax},
}

func clampReferenceBits(v, lo, hi float32) uint32 {
	if v < lo {
		return math.Float32bits(lo)
	}
	if v > hi {
		return math.Float32bits(hi)
	}
	return math.Float32bits(v)
}

func TestClampExactContract(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(804, 1104))
	values := []uint32{0, 0x80000000, 0x3f800000, 0xbf800000, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 0xff812345, 1, 0x80000001, 0x7f7fffff, 0xff7fffff}
	for i := 0; i < 100; i++ {
		values = append(values, rng.Uint32())
	}
	negzero := math.Float32frombits(0x80000000)
	bounds := [][2]float32{{-1, 1}, {-1.5, 2.5}, {0, 0}, {negzero, 0}, {0, negzero}, {negzero, negzero}, {1, 1}, {float32(math.Inf(-1)), float32(math.Inf(1))}, {float32(math.Inf(-1)), 0}, {0, float32(math.Inf(1))}, {float32(math.Inf(1)), float32(math.Inf(1))}, {float32(math.Inf(-1)), float32(math.Inf(-1))}, {-math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat32}, {-math.MaxFloat32, math.MaxFloat32}}
	for _, bound := range bounds {
		lo, hi := bound[0], bound[1]
		for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
			for _, k := range clampKernels {
				for _, alias := range []bool{false, true} {
					x := make([]float32, n)
					before := make([]uint32, n)
					for i := range x {
						before[i] = values[i%len(values)]
						x[i] = math.Float32frombits(before[i])
					}
					guard := make([]float32, n+2)
					for i := range guard {
						guard[i] = math.Float32frombits(0xdeadbeef)
					}
					out := guard[1 : n+1 : n+1]
					if alias {
						out = x
					}
					if err := k.run(x, out, lo, hi); err != nil {
						t.Fatal(k.name, err)
					}
					for i, v := range out {
						want := clampReferenceBits(math.Float32frombits(before[i]), lo, hi)
						if math.Float32bits(v) != want {
							t.Fatalf("%s n=%d alias=%v lo=%08x hi=%08x i=%d got=%08x want=%08x", k.name, n, alias, math.Float32bits(lo), math.Float32bits(hi), i, math.Float32bits(v), want)
						}
						if !alias && math.Float32bits(x[i]) != before[i] {
							t.Fatal("input changed")
						}
					}
					if !alias && (math.Float32bits(guard[0]) != 0xdeadbeef || math.Float32bits(guard[n+1]) != 0xdeadbeef) {
						t.Fatal("clamp canary")
					}
				}
			}
		}
	}
}

func TestClampErrorsRepairsAndAllocations(t *testing.T) {
	negzero := math.Float32frombits(0x80000000)
	for _, k := range clampKernels {
		for _, bound := range [][2]float32{{2, 1}, {float32(math.NaN()), 1}, {0, float32(math.NaN())}} {
			out := []float32{math.Float32frombits(0x7f812345)}
			if err := k.run([]float32{0}, out, bound[0], bound[1]); !errors.Is(err, patterns.ErrInvalidBounds) || math.Float32bits(out[0]) != 0x7f812345 {
				t.Fatal("invalid bounds wrote output", k.name, err)
			}
			if err := k.run(nil, nil, bound[0], bound[1]); !errors.Is(err, patterns.ErrInvalidBounds) {
				t.Fatal("empty invalid bounds")
			}
		}
		if err := k.run(nil, []float32{7}, 2, 1); !errors.Is(err, patterns.ErrLengthMismatch) {
			t.Fatal("length precedence")
		}
		if err := k.run(nil, nil, -1, 1); err != nil {
			t.Fatal(err)
		}
		x := []float32{-1, 1, negzero, 0, math.Float32frombits(0x7f812345)}
		out := make([]float32, len(x))
		if err := k.run(x, out, 0, negzero); err != nil {
			t.Fatal(err)
		}
		want := []uint32{0, 0x80000000, 0x80000000, 0, 0x7f812345}
		for i, v := range out {
			if math.Float32bits(v) != want[i] {
				t.Fatal("zero/NaN repair", k.name, i)
			}
		}
		if a := testing.AllocsPerRun(20, func() {
			if err := k.run(x, out, -1, 1); err != nil {
				panic(err)
			}
		}); a != 0 {
			t.Fatal("clamp allocation", k.name, a)
		}
	}
}
