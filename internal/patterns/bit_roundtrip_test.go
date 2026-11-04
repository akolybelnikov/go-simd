package patterns_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
	"math/rand/v2"
	"simd"
	"testing"
)

func TestBitRoundTripPreservesOpaqueBits(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(27, 2026))
	bits := []uint32{0, 0x80000000, 1, 0x80000001, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 0xffc54321, 0xff812345, 0x7f7fffff}
	for i := 0; i < 4096; i++ {
		bits = append(bits, rng.Uint32())
	}
	for _, kernel := range []func([]float32, []float32) error{patterns.BitRoundTripScalar, patterns.BitRoundTripSIMD} {
		for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 4107} {
			for _, alias := range []bool{false, true} {
				x := make([]float32, n)
				before := make([]uint32, n)
				for i := range x {
					before[i] = bits[i%len(bits)]
					x[i] = math.Float32frombits(before[i])
				}
				guard := make([]float32, n+2)
				guard[0], guard[n+1] = 123, 456
				out := guard[1 : n+1 : n+1]
				if alias {
					out = x
				}
				if e := kernel(x, out); e != nil {
					t.Fatal(e)
				}
				for i, v := range out {
					if math.Float32bits(v) != before[i] {
						t.Fatalf("n=%d alias=%v i=%d got=%08x want=%08x", n, alias, i, math.Float32bits(v), before[i])
					}
					if !alias && math.Float32bits(x[i]) != before[i] {
						t.Fatal("input changed")
					}
				}
				if guard[0] != 123 || guard[n+1] != 456 {
					t.Fatal("canary")
				}
			}
		}
		out := []float32{99}
		if e := kernel(nil, out); !errors.Is(e, patterns.ErrLengthMismatch) || out[0] != 99 {
			t.Fatal(e)
		}
		x, out := []float32{math.Float32frombits(0x7f812345)}, make([]float32, 1)
		if count := testing.AllocsPerRun(20, func() {
			if e := kernel(x, out); e != nil {
				panic(e)
			}
		}); count != 0 {
			t.Fatal(count)
		}
	}
}
