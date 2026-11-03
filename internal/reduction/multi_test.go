package reduction_test

import (
	"errors"
	"math"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/reduction"
)

// Software oracle: exact rounded additions, vector ordinal modulo count, then
// lane-wise accumulator merge. This is independent of the unrolled kernels.
func referenceMultiFold(x []float32, width, count int) float32 {
	accumulators := make([][]float32, count)
	for a := range accumulators {
		accumulators[a] = make([]float32, width)
	}
	full := len(x) / width
	for v := 0; v < full; v++ {
		for lane := 0; lane < width; lane++ {
			a := v % count
			accumulators[a][lane] = referenceAdd(accumulators[a][lane], x[v*width+lane])
		}
	}
	var total float32
	for lane := 0; lane < width; lane++ {
		merged := accumulators[0][lane]
		for a := 1; a < count; a++ {
			merged = referenceAdd(merged, accumulators[a][lane])
		}
		total = referenceAdd(total, merged)
	}
	for _, value := range x[full*width:] {
		total = referenceAdd(total, value)
	}
	return total
}

func TestMultiAccumulatorSpecifiedFolds(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(2026, 1103))
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 3*w - 1, 3 * w, 3*w + 1, 4*w - 1, 4 * w, 4*w + 1, 5*w + 1, 8*w - 1, 8 * w, 8*w + 1, 1025} {
		for _, kind := range []string{"random", "zeros", "special", "overflow", "subnormal", "cancellation"} {
			x, y, products := make([]float32, n), make([]float32, n), make([]float32, n)
			for i := range x {
				switch kind {
				case "random":
					x[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
					y[i] = float32(math.Ldexp(float64(int(rng.Uint32()%2001)-1000), int(rng.Uint32()%41)-20))
				case "zeros":
					x[i], y[i] = math.Float32frombits(0x80000000), 1
				case "special":
					values := [8]float32{float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), 0, -1, 1, math.MaxFloat32, math.SmallestNonzeroFloat32}
					x[i], y[i] = values[i%8], float32(i%5-2)
				case "overflow":
					x[i], y[i] = math.MaxFloat32, 2
					if i%2 == 1 {
						x[i] = -x[i]
					}
				case "subnormal":
					x[i], y[i] = math.SmallestNonzeroFloat32, float32(i%7-3)/2
				case "cancellation":
					y[i] = 1
					if i == 0 {
						x[i] = 16777216
					}
					if i == w {
						x[i] = 1
					}
					if i == 2*w {
						x[i] = -16777216
					}
				}
				products[i] = referenceMultiply(x[i], y[i])
			}
			xBefore, yBefore := append([]float32(nil), x...), append([]float32(nil), y...)
			for _, kernel := range []struct {
				count int
				sum   func([]float32) float32
				dot   func([]float32, []float32) (float32, error)
			}{
				{2, reduction.SumSIMD2, reduction.DotSIMD2}, {4, reduction.SumSIMD4, reduction.DotSIMD4},
			} {
				gotSum := kernel.sum(x)
				gotDot, err := kernel.dot(x, y)
				if !sameResult(gotSum, referenceMultiFold(x, w, kernel.count)) || err != nil || !sameResult(gotDot, referenceMultiFold(products, w, kernel.count)) {
					t.Fatalf("wrong count=%d/%s/n=%d: sum=%08x dot=%08x err=%v", kernel.count, kind, n, math.Float32bits(gotSum), math.Float32bits(gotDot), err)
				}
				if kind == "random" {
					assertForwardErrorBound(t, x, gotSum, n+w+kernel.count+1)
					assertDotForwardBound(t, x, y, gotDot, n+w+kernel.count+2)
				}
			}
			for i := range x {
				if math.Float32bits(x[i]) != math.Float32bits(xBefore[i]) || math.Float32bits(y[i]) != math.Float32bits(yBefore[i]) {
					t.Fatal("input modified")
				}
			}
		}
	}
}

func TestMultiAccumulatorRoundingAndContracts(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	for _, kernel := range []struct {
		count int
		sum   func([]float32) float32
		dot   func([]float32, []float32) (float32, error)
	}{
		{2, reduction.SumSIMD2, reduction.DotSIMD2}, {4, reduction.SumSIMD4, reduction.DotSIMD4},
	} {
		if math.Float32bits(kernel.sum(nil)) != 0 {
			t.Fatal("empty sum not +0")
		}
		if got, err := kernel.dot(nil, nil); err != nil || math.Float32bits(got) != 0 {
			t.Fatal("empty dot not +0,nil")
		}
		for _, pair := range [][2][]float32{{nil, {1}}, {{1}, nil}, {{1, 2}, {3}}} {
			if got, err := kernel.dot(pair[0], pair[1]); !errors.Is(err, reduction.ErrLengthMismatch) || math.Float32bits(got) != 0 {
				t.Fatal("length mismatch contract")
			}
		}
		shared := []float32{1, 2, 3, 4}
		if got, err := kernel.dot(shared[:3], shared[1:]); err != nil || got != 20 {
			t.Fatal("read-only overlap lost")
		}
		for _, n := range []int{kernel.count*w + 1, 2 * kernel.count * w} {
			x, y := make([]float32, n), make([]float32, n)
			x[0], y[0], x[kernel.count*w], y[kernel.count*w] = -1, 1, 1+0x1p-23, 1-0x1p-23
			if got, err := kernel.dot(x, y); err != nil || math.Float32bits(got) != 0 {
				t.Fatal("multiply/add fused in vector or tail")
			}
		}
		for _, layout := range []int{1, 2} {
			x, y := make([]float32, 8*w), make([]float32, 8*w)
			for i := range y {
				y[i] = 1
			}
			x[0], x[layout*w], x[2*layout*w] = 16777216, 1, -16777216
			want := float32(0)
			if kernel.count == 2 && layout == 1 || kernel.count == 4 && layout == 2 {
				want = 1
			}
			got, err := kernel.dot(x, y)
			if kernel.sum(x) != want || err != nil || got != want {
				t.Fatal("accumulator-order counterexample lost")
			}
		}
		if w <= 16 {
			x, y := make([]float32, 1025), make([]float32, 1025)
			if allocations := testing.AllocsPerRun(20, func() { retained = kernel.sum(x) }); allocations != 0 {
				t.Fatal("sum allocations", allocations)
			}
			if allocations := testing.AllocsPerRun(20, func() {
				var err error
				retained, err = kernel.dot(x, y)
				if err != nil {
					panic(err)
				}
			}); allocations != 0 {
				t.Fatal("dot allocations", allocations)
			}
		}
	}
}
