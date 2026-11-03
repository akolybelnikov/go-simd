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

// Independent exact multiply/add followed by one binary32 rounding. Do not
// narrow a binary64 FMA: that can double round at binary32 midpoints.
func referenceFMAUpdate(a, b, c float32) float32 {
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
	product := new(big.Rat).Mul(new(big.Rat).SetFloat64(float64(a)), new(big.Rat).SetFloat64(float64(b)))
	sum := new(big.Rat).Add(product, new(big.Rat).SetFloat64(float64(c)))
	if sum.Sign() == 0 {
		if (a == 0 || b == 0) && (math.Signbit(float64(a)) != math.Signbit(float64(b))) && c == 0 && math.Signbit(float64(c)) {
			return math.Float32frombits(0x80000000)
		}
		return 0
	}
	value, _ := sum.Float32()
	return value
}

func referenceFusedDot(x, y []float32, width, count int) float32 {
	var total float32
	if count == 0 {
		for i, v := range x {
			total = referenceFMAUpdate(v, y[i], total)
		}
		return total
	}
	acc := make([][]float32, count)
	for a := range acc {
		acc[a] = make([]float32, width)
	}
	full := len(x) / width
	for v := 0; v < full; v++ {
		for lane := 0; lane < width; lane++ {
			i := v*width + lane
			acc[v%count][lane] = referenceFMAUpdate(x[i], y[i], acc[v%count][lane])
		}
	}
	for lane := 0; lane < width; lane++ {
		merged := acc[0][lane]
		for a := 1; a < count; a++ {
			merged = referenceAdd(merged, acc[a][lane])
		}
		total = referenceAdd(total, merged)
	}
	for i := full * width; i < len(x); i++ {
		total = referenceFMAUpdate(x[i], y[i], total)
	}
	return total
}

var fusedDotKernels = []struct {
	count int
	run   func([]float32, []float32) (float32, error)
}{
	{0, reduction.DotFMAScalar}, {1, reduction.DotFMASIMD}, {2, reduction.DotFMASIMD2}, {4, reduction.DotFMASIMD4},
}

func TestFusedDotSpecifiedOrders(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(419, 1103))
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 3*w + 1, 4*w - 1, 4 * w, 4*w + 1, 5*w + 1, 8*w + 1, 1025} {
		for _, kind := range []string{"finite", "bits", "negative-zero", "underflow", "special", "fusion"} {
			x, y := make([]float32, n), make([]float32, n)
			for i := range x {
				switch kind {
				case "finite":
					x[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
					y[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
				case "bits":
					x[i], y[i] = math.Float32frombits(rng.Uint32()), math.Float32frombits(rng.Uint32())
				case "negative-zero":
					x[i], y[i] = math.Float32frombits(0x80000000), 1
				case "underflow":
					x[i], y[i] = math.SmallestNonzeroFloat32, -0.25
				case "special":
					v := [8]float32{math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), 0, math.Float32frombits(0x7fc12345), math.SmallestNonzeroFloat32, 1}
					x[i], y[i] = v[i%8], float32(i%5-2)
				case "fusion":
					if i == 0 {
						x[i], y[i] = -1, 1
					}
					if i == w {
						x[i], y[i] = 1+0x1p-23, 1-0x1p-23
					}
				}
			}
			xBefore, yBefore := append([]float32(nil), x...), append([]float32(nil), y...)
			for _, kernel := range fusedDotKernels {
				got, err := kernel.run(x, y)
				want := referenceFusedDot(x, y, w, kernel.count)
				if err != nil || !sameResult(got, want) {
					t.Fatalf("count=%d/%s/n=%d: %08x != %08x (%v)", kernel.count, kind, n, math.Float32bits(got), math.Float32bits(want), err)
				}
				if kind == "finite" {
					assertDotForwardBound(t, x, y, got, n+w+kernel.count+1)
				}
			}
			for i := range x {
				if math.Float32bits(x[i]) != math.Float32bits(xBefore[i]) || math.Float32bits(y[i]) != math.Float32bits(yBefore[i]) {
					t.Fatal("fused dot modified input")
				}
			}
		}
	}
}

func TestFusedDotContractsAndDistinguishingCases(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	for _, kernel := range fusedDotKernels {
		if got, err := kernel.run(nil, nil); err != nil || math.Float32bits(got) != 0 {
			t.Fatal("empty fused dot must be +0,nil")
		}
		for _, pair := range [][2][]float32{{nil, {1}}, {{1}, nil}, {{1, 2}, {3}}} {
			if got, err := kernel.run(pair[0], pair[1]); !errors.Is(err, reduction.ErrLengthMismatch) || math.Float32bits(got) != 0 {
				t.Fatal("length contract lost")
			}
		}
		backing := make([]float32, 2*w+2)
		for i := range backing {
			backing[i] = float32(i%7 - 3)
		}
		x, y := backing[:2*w+1], backing[1:]
		want := referenceFusedDot(x, y, w, kernel.count)
		if got, err := kernel.run(x, y); err != nil || !sameResult(got, want) {
			t.Fatal("read-only overlap failed")
		}
		if w <= 16 {
			if allocations := testing.AllocsPerRun(15, func() {
				var err error
				retained, err = kernel.run(x, y)
				if err != nil {
					panic(err)
				}
			}); allocations != 0 {
				t.Fatal("fused dot allocations", allocations)
			}
		}
	}
	x, y := make([]float32, 2*w), make([]float32, 2*w)
	x[0], y[0], x[w], y[w] = -1, 1, 1+0x1p-23, 1-0x1p-23
	for _, kernel := range fusedDotKernels {
		got, err := kernel.run(x, y)
		want := float32(0)
		if kernel.count <= 1 {
			want = -0x1p-46
		}
		if err != nil || math.Float32bits(got) != math.Float32bits(want) {
			t.Fatal("fusion/fold distinction lost", kernel.count, got)
		}
	}
	if got, _ := reduction.DotScalar(x, y); got != 0 {
		t.Fatal("old dot unexpectedly fused")
	}
	x[0], y[0], x[w], y[w] = -math.MaxFloat32, 1, math.MaxFloat32, 2
	for _, kernel := range fusedDotKernels {
		got, err := kernel.run(x, y)
		if err != nil || kernel.count <= 1 && got != math.MaxFloat32 || kernel.count > 1 && !math.IsInf(float64(got), 1) {
			t.Fatal("overflow/fold distinction lost", kernel.count, got)
		}
	}
	x, y = make([]float32, w), make([]float32, w)
	for i := range x {
		x[i], y[i] = math.SmallestNonzeroFloat32, -0.25
	}
	scalar, _ := reduction.DotFMAScalar(x, y)
	striped, _ := reduction.DotFMASIMD(x, y)
	if math.Float32bits(scalar) != 0x80000000 || math.Float32bits(striped) != 0 {
		t.Fatal("signed underflow/fold example lost")
	}
}
