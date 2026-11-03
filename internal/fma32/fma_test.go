package fma32_test

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/fma32"
)

// Exact rational multiply/add and one binary32 rounding, independent of TwoSum.
func oracle(a, b, c float32) float32 {
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

func same(a, b float32) bool {
	return math.Float32bits(a) == math.Float32bits(b) || math.IsNaN(float64(a)) && math.IsNaN(float64(b))
}

func TestFMAAgainstExactRationalOracle(t *testing.T) {
	u := float32(0x1p-23)
	cases := [][3]float32{
		{math.MaxFloat32, 2, -math.MaxFloat32}, // Fused arithmetic rescues intermediate overflow.
		{1 + u, 1 - u, -1}, {1 + u, 1 - u, 16777218},
		{float32(math.Ldexp(float64(1+u), -75)), float32(math.Ldexp(float64(1-u), -75)), math.Float32frombits(0x007fffff)},
		{float32(math.Ldexp(float64(1+u), 51)), float32(math.Ldexp(float64(1-u), 52)), math.MaxFloat32},
	}
	values := []float32{0, math.Float32frombits(0x80000000), 1, -1, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, math.Float32frombits(0x00800000), math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345)}
	for _, a := range values {
		for _, b := range values {
			for _, c := range values {
				cases = append(cases, [3]float32{a, b, c})
			}
		}
	}
	rng := rand.New(rand.NewPCG(2026, 1104))
	for i := 0; i < 30000; i++ {
		cases = append(cases, [3]float32{math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32())})
	}
	for _, values := range cases {
		a, b, c := values[0], values[1], values[2]
		for _, negate := range []bool{false, true} {
			if negate {
				a, c = -a, -c
			}
			got, want := fma32.FMA(a, b, c), oracle(a, b, c)
			if !same(got, want) {
				t.Fatalf("a=%08x b=%08x c=%08x: got %08x want %08x", math.Float32bits(a), math.Float32bits(b), math.Float32bits(c), math.Float32bits(got), math.Float32bits(want))
			}
		}
	}
}

func TestWidenedFMAIsNotAFloat32Oracle(t *testing.T) {
	u := float32(0x1p-23)
	for _, v := range [][3]float32{
		{1 + u, 1 - u, 16777218},
		{float32(math.Ldexp(float64(1+u), -75)), float32(math.Ldexp(float64(1-u), -75)), math.Float32frombits(0x007fffff)},
		{float32(math.Ldexp(float64(1+u), 51)), float32(math.Ldexp(float64(1-u), 52)), math.MaxFloat32},
	} {
		want := oracle(v[0], v[1], v[2])
		if same(float32(math.FMA(float64(v[0]), float64(v[1]), float64(v[2]))), want) {
			t.Fatal("double-rounding counterexample lost")
		}
		if !same(fma32.FMA(v[0], v[1], v[2]), want) {
			t.Fatal("midpoint repair failed")
		}
	}
}
