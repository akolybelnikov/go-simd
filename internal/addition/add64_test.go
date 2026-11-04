package addition_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/addition"
	"math"
	"math/big"
	"math/rand/v2"
	"simd"
	"testing"
)

func add64Oracle(a, b float64) float64 {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.NaN()
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		if math.IsInf(a, 1) && math.IsInf(b, -1) || math.IsInf(a, -1) && math.IsInf(b, 1) {
			return math.NaN()
		}
		if math.IsInf(a, 0) {
			return a
		}
		return b
	}
	if a == 0 && b == 0 && math.Signbit(a) && math.Signbit(b) {
		return math.Float64frombits(1 << 63)
	}
	value := new(big.Rat).Add(new(big.Rat).SetFloat64(a), new(big.Rat).SetFloat64(b))
	out, _ := value.Float64()
	return out
}
func TestAddition64IndependentBitsTailsAndAliases(t *testing.T) {
	var vector simd.Float64s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(64, 1104))
	nz := math.Float64frombits(1 << 63)
	cases := [][2]float64{{1, 0x1p-53}, {1 + 0x1p-52, 0x1p-53}, {nz, nz}, {nz, 0}, {math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64}, {math.MaxFloat64, math.MaxFloat64}, {math.Inf(1), math.Inf(-1)}, {math.Float64frombits(0x7ff0000000012345), 1}}
	for i := 0; i < 2048; i++ {
		cases = append(cases, [2]float64{math.Float64frombits(rng.Uint64()), math.Float64frombits(rng.Uint64())})
	}
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 2057} {
		for _, kernel := range []func([]float64, []float64, []float64) error{addition.Add64Scalar, addition.Add64SIMD} {
			for alias := -1; alias < 2; alias++ {
				a, b := make([]float64, n), make([]float64, n)
				before := make([][2]uint64, n)
				want := make([]float64, n)
				for i := range a {
					p := cases[i%len(cases)]
					a[i], b[i] = p[0], p[1]
					before[i] = [2]uint64{math.Float64bits(a[i]), math.Float64bits(b[i])}
					want[i] = add64Oracle(a[i], b[i])
				}
				guard := make([]float64, n+2)
				guard[0] = 123
				guard[n+1] = 456
				out := guard[1 : n+1 : n+1]
				if alias >= 0 {
					out = [][]float64{a, b}[alias]
				}
				if err := kernel(a, b, out); err != nil {
					t.Fatal(err)
				}
				for i, v := range out {
					if math.Float64bits(v) != math.Float64bits(want[i]) && !(math.IsNaN(v) && math.IsNaN(want[i])) {
						t.Fatalf("n=%d alias=%d i=%d got=%016x want=%016x", n, alias, i, math.Float64bits(v), math.Float64bits(want[i]))
					}
					if alias < 0 && before[i] != ([2]uint64{math.Float64bits(a[i]), math.Float64bits(b[i])}) {
						t.Fatal("input changed")
					}
				}
				if guard[0] != 123 || guard[n+1] != 456 {
					t.Fatal("canary")
				}
			}
		}
	}
	for _, kernel := range []func([]float64, []float64, []float64) error{addition.Add64Scalar, addition.Add64SIMD} {
		for short := 0; short < 3; short++ {
			arrays := [][]float64{{1}, {2}, {99}}
			arrays[short] = nil
			if e := kernel(arrays[0], arrays[1], arrays[2]); !errors.Is(e, addition.ErrLengthMismatch) {
				t.Fatal(e)
			}
			if short != 2 && arrays[2][0] != 99 {
				t.Fatal("error wrote output")
			}
		}
		x, out := []float64{1, 2, 3}, make([]float64, 3)
		if a := testing.AllocsPerRun(20, func() {
			if e := kernel(x, x, out); e != nil {
				panic(e)
			}
		}); a != 0 {
			t.Fatal(a)
		}
	}
}
