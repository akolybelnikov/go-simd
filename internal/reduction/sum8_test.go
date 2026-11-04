package reduction_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"math"
	"math/big"
	"math/rand/v2"
	"simd"
	"testing"
)

func sum8OracleAdd(a, b float32) float32 {
	if math.IsNaN(float64(a)) || math.IsNaN(float64(b)) {
		return float32(math.NaN())
	}
	if math.IsInf(float64(a), 0) || math.IsInf(float64(b), 0) {
		if math.IsInf(float64(a), 1) && math.IsInf(float64(b), -1) || math.IsInf(float64(a), -1) && math.IsInf(float64(b), 1) {
			return float32(math.NaN())
		}
		if math.IsInf(float64(a), 0) {
			return a
		}
		return b
	}
	if a == 0 && b == 0 && math.Signbit(float64(a)) && math.Signbit(float64(b)) {
		return math.Float32frombits(0x80000000)
	}
	value := new(big.Rat).Add(new(big.Rat).SetFloat64(float64(a)), new(big.Rat).SetFloat64(float64(b)))
	out, _ := value.Float32()
	return out
}
func sum8Oracle(x []float32) []float32 {
	count := len(x) / 8
	if len(x)%8 != 0 {
		count++
	}
	out := make([]float32, count)
	for block := range count {
		var p [8]float32
		copy(p[:], x[block*8:])
		var pairs [4]float32
		for i := range 4 {
			pairs[i] = sum8OracleAdd(p[i], p[i+4])
		}
		value := pairs[0]
		for i := 1; i < 4; i++ {
			value = sum8OracleAdd(value, pairs[i])
		}
		out[block] = value
	}
	return out
}
func TestSum8IndependentFoldExtentsAndAliases(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(29, 2026))
	values := []uint32{0, 0x80000000, 0x3f800000, 0xbf800000, 1, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 0x7f7fffff}
	for i := 0; i < 2048; i++ {
		values = append(values, rng.Uint32())
	}
	for _, kernel := range []func([]float32, []float32) error{reduction.Sum8Scalar, reduction.Sum8SIMD} {
		for _, n := range []int{0, 1, 3, 4, 5, 7, 8, 9, 15, 16, 17, w - 1, w, w + 1, 2*w + 1, 2057} {
			for _, alias := range []bool{false, true} {
				x := make([]float32, n)
				before := make([]uint32, n)
				for i := range x {
					before[i] = values[i%len(values)]
					x[i] = math.Float32frombits(before[i])
				}
				want := sum8Oracle(x)
				guard := make([]float32, len(want)+2)
				guard[0], guard[len(guard)-1] = 123, 456
				out := guard[1 : len(guard)-1 : len(guard)-1]
				if alias {
					out = x[:len(want)]
				}
				if e := kernel(x, out); e != nil {
					t.Fatal(e)
				}
				for i, v := range out {
					if math.Float32bits(v) != math.Float32bits(want[i]) && !(math.IsNaN(float64(v)) && math.IsNaN(float64(want[i]))) {
						t.Fatalf("n=%d alias=%v out=%d got=%08x want=%08x", n, alias, i, math.Float32bits(v), math.Float32bits(want[i]))
					}
				}
				if !alias {
					for i, v := range x {
						if math.Float32bits(v) != before[i] {
							t.Fatal("input changed")
						}
					}
				}
				if guard[0] != 123 || guard[len(guard)-1] != 456 {
					t.Fatal("canary")
				}
			}
		}
		out := []float32{99}
		if e := kernel(make([]float32, 9), out); !errors.Is(e, reduction.ErrLengthMismatch) || out[0] != 99 {
			t.Fatal(e)
		}
		if e := kernel(nil, out); !errors.Is(e, reduction.ErrLengthMismatch) || out[0] != 99 {
			t.Fatal(e)
		}
		x, out := []float32{-1e20, 1, 1e20, 0, 1e20, 0, -1e20, 0, 2}, make([]float32, 2)
		if e := kernel(x, out); e != nil || out[0] != 1 || out[1] != 2 {
			t.Fatal("specified pair fold", out, e)
		}
		if count := testing.AllocsPerRun(20, func() {
			if e := kernel(x, out); e != nil {
				panic(e)
			}
		}); count != 0 {
			t.Fatal(count)
		}
	}
}
