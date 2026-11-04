package movement

import (
	"errors"
	"math"
	"math/rand/v2"
	"simd"
	"testing"
)

func TestCopy64RawRepresentations(t *testing.T) {
	var vector simd.Float64s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(52, 91))
	for _, fn := range []func([]float64, []float64) error{Copy64Scalar, Copy64Go, Copy64SIMD} {
		for _, n := range []int{0, 1, 2, 3, w - 1, w, w + 1, 2*w + 1, 1025} {
			x := make([]float64, n)
			want := make([]uint64, n)
			special := []uint64{0, 1 << 63, 1, 0x8000000000000001, 0x7ff0000000000000, 0xfff0000000000000, 0x7ff8123456789abc, 0xfff8123456789abc, 0x7ff0123456789abc, 0xfff0123456789abc}
			for i := range x {
				bits := rng.Uint64()
				if i < len(special) {
					bits = special[i]
				}
				x[i] = math.Float64frombits(bits)
				want[i] = bits
			}
			backing := make([]float64, n+2)
			backing[0] = 123
			backing[n+1] = 456
			out := backing[1 : n+1]
			if e := fn(x, out); e != nil {
				t.Fatal(e)
			}
			if backing[0] != 123 || backing[n+1] != 456 {
				t.Fatal("canary")
			}
			for i := range x {
				if math.Float64bits(x[i]) != want[i] || math.Float64bits(out[i]) != want[i] {
					t.Fatal("raw bits/input preservation", n, i)
				}
			}
			if e := fn(x, x); e != nil {
				t.Fatal(e)
			}
			for i := range x {
				if math.Float64bits(x[i]) != want[i] {
					t.Fatal("exact alias", i)
				}
			}
		}
		out := []float64{99, 99}
		if e := fn([]float64{1}, out); !errors.Is(e, ErrLengthMismatch) || out[0] != 99 || out[1] != 99 {
			t.Fatal("error before write", e)
		}
		x, y := make([]float64, 17), make([]float64, 17)
		if got := testing.AllocsPerRun(10, func() {
			if e := fn(x, y); e != nil {
				panic(e)
			}
		}); got != 0 {
			t.Fatal("allocation", got)
		}
	}
}
