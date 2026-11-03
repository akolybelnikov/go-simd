package reduction_test

import (
	"errors"
	"math"
	"math/big"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

// Exact finite multiplication followed by a single float32 rounding, independent
// of the kernel's multiply/add expression. Special values preserve IEEE signs.
func referenceMultiply(a, b float32) float32 {
	if a == 0 || b == 0 || math.IsNaN(float64(a)) || math.IsNaN(float64(b)) || math.IsInf(float64(a), 0) || math.IsInf(float64(b), 0) {
		return float32(float64(a) * float64(b))
	}
	x, y := new(big.Rat).SetFloat64(float64(a)), new(big.Rat).SetFloat64(float64(b))
	value, _ := new(big.Rat).Mul(x, y).Float32()
	return value
}

func TestSpecifiedUnfusedDotFolds(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(938, 271))
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 65, 1025} {
		for _, kind := range []string{"random", "zeros", "special", "cancellation", "overflow", "subnormal"} {
			x, y, products := make([]float32, n), make([]float32, n), make([]float32, n)
			for i := range x {
				switch kind {
				case "random":
					x[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
					y[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
				case "zeros":
					x[i], y[i] = math.Float32frombits(0x80000000), 1
				case "special":
					xv := [8]float32{float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), 0, -1, 1, math.MaxFloat32, math.SmallestNonzeroFloat32}
					yv := [8]float32{0, -1, 1, float32(math.Inf(1)), -1, 1, 2, 0.5}
					x[i], y[i] = xv[i%8], yv[i%8]
				case "cancellation":
					y[i] = 1
					switch i % 64 {
					case 0:
						x[i] = 16777216
					case 1:
						x[i] = 1
					case 32:
						x[i] = -16777216
					}
				case "overflow":
					x[i], y[i] = math.MaxFloat32, 2
					if i%2 == 1 {
						y[i] = -2
					}
				case "subnormal":
					x[i], y[i] = math.SmallestNonzeroFloat32, float32(i%7-3)/2
				}
				products[i] = referenceMultiply(x[i], y[i])
			}
			xBefore, yBefore := append([]float32(nil), x...), append([]float32(nil), y...)
			for _, kernel := range []struct {
				name  string
				width int
				run   func([]float32, []float32) (float32, error)
			}{
				{"scalar", 1, reduction.DotScalar}, {"SIMD", w, reduction.DotSIMD},
			} {
				got, err := kernel.run(x, y)
				want := referenceFold(products, kernel.width)
				if err != nil || !sameResult(got, want) {
					t.Fatalf("%s/%s/n=%d: got %08x want %08x err=%v", kernel.name, kind, n, math.Float32bits(got), math.Float32bits(want), err)
				}
				if kind == "random" {
					assertDotForwardBound(t, x, y, got, n+w+2)
				}
			}
			for i := range x {
				if math.Float32bits(x[i]) != math.Float32bits(xBefore[i]) || math.Float32bits(y[i]) != math.Float32bits(yBefore[i]) {
					t.Fatal("dot modified input")
				}
			}
		}
	}
}

// For finite normal-range random cases, include multiplication and addition
// rounding in a conservative absolute bound against exact mathematical x*y.
func assertDotForwardBound(t *testing.T, x, y []float32, got float32, k int) {
	t.Helper()
	exact, norm := new(big.Rat), new(big.Rat)
	for i, v := range x {
		term := new(big.Rat).Mul(new(big.Rat).SetFloat64(float64(v)), new(big.Rat).SetFloat64(float64(y[i])))
		exact.Add(exact, term)
		norm.Add(norm, new(big.Rat).Abs(term))
	}
	difference := new(big.Rat).Sub(new(big.Rat).SetFloat64(float64(got)), exact)
	difference.Abs(difference)
	bound := new(big.Rat).Mul(norm, big.NewRat(int64(k), (1<<24)-int64(k)))
	if difference.Cmp(bound) > 0 {
		t.Fatalf("dot exceeds forward bound: %s > %s", difference, bound)
	}
}

func TestDotUnfusedRoundingOrderAndLengthContract(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	for _, kernel := range []func([]float32, []float32) (float32, error){reduction.DotScalar, reduction.DotSIMD} {
		for _, pair := range [][2][]float32{{nil, {1}}, {{1}, nil}, {{1, 2}, {3}}} {
			got, err := kernel(pair[0], pair[1])
			if !errors.Is(err, reduction.ErrLengthMismatch) || math.Float32bits(got) != 0 {
				t.Fatal("length mismatch contract lost")
			}
		}
		if got, err := kernel(nil, nil); err != nil || math.Float32bits(got) != 0 {
			t.Fatal("empty dot must return +0,nil")
		}
		// A fused second multiply-add would retain -2^-46; unfused returns 0.
		for _, n := range []int{w + 1, 2 * w} {
			x, y := make([]float32, n), make([]float32, n)
			x[0], y[0], x[w], y[w] = -1, 1, 1+0x1p-23, 1-0x1p-23
			if got, err := kernel(x, y); err != nil || math.Float32bits(got) != 0 {
				t.Fatal("product/add unexpectedly fused")
			}
		}
		shared := []float32{1, 2, 3, 4}
		if got, err := kernel(shared[:3], shared[1:]); err != nil || got != 20 {
			t.Fatal("read-only overlap rejected or wrong")
		}
	}
	x, y := make([]float32, 2*w), make([]float32, 2*w)
	for i := range y {
		y[i] = 1
	}
	x[0], x[1], x[w] = 16777216, 1, -16777216
	a, _ := reduction.DotScalar(x, y)
	b, _ := reduction.DotSIMD(x, y)
	if a != 0 || b != 1 {
		t.Fatal("dot order counterexample lost")
	}
	x[0], x[1], x[w] = 16777216, -16777216, 1
	a, _ = reduction.DotScalar(x, y)
	b, _ = reduction.DotSIMD(x, y)
	if a != 1 || b != 0 {
		t.Fatal("inverse dot order counterexample lost")
	}
}

func TestCurrentDotKernelsDoNotAllocate(t *testing.T) {
	var vector simd.Float32s
	if vector.Len() > 16 {
		t.Skip("larger future width uses declared scratch")
	}
	x, y := make([]float32, 1025), make([]float32, 1025)
	for _, kernel := range []func([]float32, []float32) (float32, error){reduction.DotScalar, reduction.DotSIMD} {
		if got := testing.AllocsPerRun(20, func() {
			var err error
			retained, err = kernel(x, y)
			if err != nil {
				panic(err)
			}
		}); got != 0 {
			t.Fatalf("dot allocations=%g", got)
		}
	}
}
