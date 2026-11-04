package patterns_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
	"math/big"
	"math/rand/v2"
	"simd"
	"testing"
)

func muladdOracle(a, b, c float32) float32 {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) || math.IsNaN(float64(c)) {
		return float32(math.NaN())
	}
	if math.IsInf(float64(a), 0) || math.IsInf(float64(b), 0) {
		if a == 0 || b == 0 {
			return float32(math.NaN())
		}
		negative := math.Signbit(float64(a)) != math.Signbit(float64(b))
		if math.IsInf(float64(c), 0) && math.Signbit(float64(c)) != negative {
			return float32(math.NaN())
		}
		if negative {
			return float32(math.Inf(-1))
		}
		return float32(math.Inf(1))
	}
	if math.IsInf(float64(c), 0) {
		return c
	}
	p := new(big.Rat).Mul(new(big.Rat).SetFloat64(float64(a)), new(big.Rat).SetFloat64(float64(b)))
	s := new(big.Rat).Add(p, new(big.Rat).SetFloat64(float64(c)))
	if s.Sign() == 0 {
		if (a == 0 || b == 0) && (math.Signbit(float64(a)) != math.Signbit(float64(b))) && c == 0 && math.Signbit(float64(c)) {
			return math.Float32frombits(0x80000000)
		}
		return 0
	}
	value, _ := s.Float32()
	return value
}

func TestMulAddIndependentBitsAliasesAndTails(t *testing.T) {
	var v simd.Float32s
	w := v.Len()
	rng := rand.New(rand.NewPCG(419, 2026))
	special := [][3]float32{{1 + 0x1p-23, 1 - 0x1p-23, -1}, {1 + 0x1p-23, 1 - 0x1p-23, 16777218}, {math.MaxFloat32, 2, -math.MaxFloat32}, {math.SmallestNonzeroFloat32, 0.5, math.SmallestNonzeroFloat32}, {math.Float32frombits(0x80000000), 1, math.Float32frombits(0x80000000)}, {float32(math.Inf(1)), 0, 1}, {float32(math.Inf(1)), 1, float32(math.Inf(-1))}, {math.Float32frombits(0x7f812345), 1, 2}}
	for i := 0; i < 400; i++ {
		special = append(special, [3]float32{math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32())})
	}
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
		for _, kernel := range []func([]float32, []float32, []float32, []float32) error{patterns.MulAddScalar, patterns.MulAddSIMD} {
			for alias := -1; alias < 3; alias++ {
				a, b, c := make([]float32, n), make([]float32, n), make([]float32, n)
				want := make([]float32, n)
				before := make([][3]uint32, n)
				for i := range a {
					p := special[i%len(special)]
					a[i], b[i], c[i] = p[0], p[1], p[2]
					before[i] = [3]uint32{math.Float32bits(a[i]), math.Float32bits(b[i]), math.Float32bits(c[i])}
					want[i] = muladdOracle(a[i], b[i], c[i])
				}
				guard := make([]float32, n+2)
				guard[0] = 123
				guard[n+1] = 456
				out := guard[1 : n+1 : n+1]
				if alias >= 0 {
					out = [][]float32{a, b, c}[alias]
				}
				if err := kernel(a, b, c, out); err != nil {
					t.Fatal(err)
				}
				for i, value := range out {
					if math.Float32bits(value) != math.Float32bits(want[i]) && !(math.IsNaN(float64(value)) && math.IsNaN(float64(want[i]))) {
						t.Fatalf("n=%d alias=%d i=%d got=%08x want=%08x", n, alias, i, math.Float32bits(value), math.Float32bits(want[i]))
					}
					if alias < 0 && before[i] != ([3]uint32{math.Float32bits(a[i]), math.Float32bits(b[i]), math.Float32bits(c[i])}) {
						t.Fatal("input changed")
					}
				}
				if guard[0] != 123 || guard[n+1] != 456 {
					t.Fatal("output canary")
				}
			}
		}
	}
}
func TestMulAddErrorsAndAllocations(t *testing.T) {
	for _, kernel := range []func([]float32, []float32, []float32, []float32) error{patterns.MulAddScalar, patterns.MulAddSIMD} {
		for short := 0; short < 4; short++ {
			arrays := [][]float32{{1}, {2}, {3}, {99}}
			arrays[short] = nil
			if err := kernel(arrays[0], arrays[1], arrays[2], arrays[3]); !errors.Is(err, patterns.ErrLengthMismatch) {
				t.Fatal(err)
			}
			if short != 3 && arrays[3][0] != 99 {
				t.Fatal("error wrote output")
			}
		}
		if err := kernel(nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		x, out := []float32{1, 2, 3}, make([]float32, 3)
		if allocations := testing.AllocsPerRun(20, func() {
			if e := kernel(x, x, x, out); e != nil {
				panic(e)
			}
		}); allocations != 0 {
			t.Fatal(allocations)
		}
	}
}
