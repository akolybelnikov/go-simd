package reduction_test

import (
	"math"
	"math/big"
	"math/rand/v2"
	"os"
	"simd"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

// Independent IEEE float32 addition: exact rational arithmetic and one rounding.
// NaN payloads are deliberately outside the contract.
func referenceAdd(a, b float32) float32 {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		return float32(math.NaN())
	}
	if math.IsInf(float64(a), 0) || math.IsInf(float64(b), 0) {
		return float32(float64(a) + float64(b))
	}
	if a == 0 && b == 0 && math.Signbit(float64(a)) && math.Signbit(float64(b)) {
		return math.Float32frombits(0x80000000)
	}
	x, y := new(big.Rat).SetFloat64(float64(a)), new(big.Rat).SetFloat64(float64(b))
	value, _ := new(big.Rat).Add(x, y).Float32()
	return value
}

func referenceFold(x []float32, width int) float32 {
	var total float32
	if width == 1 {
		for _, v := range x {
			total = referenceAdd(total, v)
		}
		return total
	}
	lanes := make([]float32, width)
	i := 0
	for ; i+width <= len(x); i += width {
		for j := range lanes {
			lanes[j] = referenceAdd(lanes[j], x[i+j])
		}
	}
	for _, v := range lanes {
		total = referenceAdd(total, v)
	}
	for ; i < len(x); i++ {
		total = referenceAdd(total, x[i])
	}
	return total
}

func sameResult(a, b float32) bool {
	return math.Float32bits(a) == math.Float32bits(b) || (math.IsNaN(float64(a)) && math.IsNaN(float64(b)))
}

func TestSpecifiedSumFolds(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	if strings.Contains(os.Getenv("GODEBUG"), "simd=0") && !simd.Emulated() {
		t.Fatal("emulation not enabled")
	}
	rng := rand.New(rand.NewPCG(123, 456))
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 65, 1025} {
		for _, kind := range []string{"random", "zeros", "special", "cancellation", "overflow"} {
			x := make([]float32, n)
			for i := range x {
				switch kind {
				case "random":
					x[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
				case "zeros":
					x[i] = math.Float32frombits(0x80000000)
				case "special":
					v := [8]float32{math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), 0, -1, 1}
					x[i] = v[i%8]
				case "cancellation":
					switch i % 64 {
					case 0:
						x[i] = 16777216
					case 1:
						x[i] = 1
					case 32:
						x[i] = -16777216
					}
				case "overflow":
					if i%w == 0 {
						x[i] = math.MaxFloat32
					}
					if i%w == 1 {
						x[i] = -math.MaxFloat32
					}
				}
			}
			before := append([]float32(nil), x...)
			for _, kernel := range []struct {
				name  string
				width int
				run   func([]float32) float32
			}{{"scalar", 1, reduction.SumScalar}, {"SIMD", w, reduction.SumSIMD}} {
				got, want := kernel.run(x), referenceFold(x, kernel.width)
				if !sameResult(got, want) {
					t.Fatalf("%s/%s/n=%d: got 0x%08x want 0x%08x", kernel.name, kind, n, math.Float32bits(got), math.Float32bits(want))
				}
				if kind == "random" {
					assertForwardErrorBound(t, x, got, n+w+1)
				}
			}
			for i := range x {
				if math.Float32bits(x[i]) != math.Float32bits(before[i]) {
					t.Fatal("input modified")
				}
			}
		}
	}
}

// For finite, nonoverflowing cases with k*u<1, test the absolute forward bound
// against exact rational sum/norm. No fixed relative tolerance near cancellation.
func assertForwardErrorBound(t *testing.T, x []float32, got float32, k int) {
	t.Helper()
	exact, norm := new(big.Rat), new(big.Rat)
	for _, v := range x {
		term := new(big.Rat).SetFloat64(float64(v))
		exact.Add(exact, term)
		norm.Add(norm, new(big.Rat).Abs(term))
	}
	difference := new(big.Rat).Sub(new(big.Rat).SetFloat64(float64(got)), exact)
	difference.Abs(difference)
	bound := new(big.Rat).Mul(norm, big.NewRat(int64(k), (1<<24)-int64(k)))
	if difference.Cmp(bound) > 0 {
		t.Fatalf("sum exceeds absolute forward-error bound: %s > %s", difference, bound)
	}
}

func TestSumOrderCounterexampleAndEmpty(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	x := make([]float32, 2*w)
	x[0], x[1], x[w] = 16777216, 1, -16777216
	if reduction.SumScalar(x) != 0 || reduction.SumSIMD(x) != 1 {
		t.Fatal("documented order distinction lost")
	}
	for i := range x {
		x[i] = 0
	}
	x[0], x[1], x[w] = 16777216, -16777216, 1
	if reduction.SumScalar(x) != 1 || reduction.SumSIMD(x) != 0 {
		t.Fatal("inverse order distinction lost")
	}
	for _, kernel := range []func([]float32) float32{reduction.SumScalar, reduction.SumSIMD} {
		if math.Float32bits(kernel(nil)) != 0 {
			t.Fatal("empty sum is not +0")
		}
	}
	for i := range x {
		x[i] = 0
	}
	x[0], x[1], x[w], x[w+1] = math.MaxFloat32, -math.MaxFloat32, math.MaxFloat32, -math.MaxFloat32
	if reduction.SumScalar(x) != 0 || !math.IsNaN(float64(reduction.SumSIMD(x))) {
		t.Fatal("overflow order distinction lost")
	}
}

var retained float32

func TestCurrentSumKernelsDoNotAllocate(t *testing.T) {
	var vector simd.Float32s
	if vector.Len() > 16 {
		t.Skip("larger future vectors use declared scratch allocation")
	}
	x := make([]float32, 1025)
	for _, kernel := range []func([]float32) float32{reduction.SumScalar, reduction.SumSIMD} {
		if got := testing.AllocsPerRun(20, func() { retained = kernel(x) }); got != 0 {
			t.Fatalf("sum allocations=%g", got)
		}
	}
}
