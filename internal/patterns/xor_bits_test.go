package patterns_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
	"math/rand/v2"
	"simd"
	"testing"
)

func xorTruthBits(a, b uint32) uint32 {
	var out uint32
	for bit := uint(0); bit < 32; bit++ {
		if (a>>bit)&1 != (b>>bit)&1 {
			out |= 1 << bit
		}
	}
	return out
}
func TestXorBitsTruthTailsAndAliases(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	rng := rand.New(rand.NewPCG(28, 2026))
	pairs := [][2]uint32{{0x3f800000, 0x80000000}, {0x7f812345, 0}, {0x7fc12345, 0xffffffff}, {0x80000000, 0x80000000}, {1, 0xffffffff}, {0x7f800000, 0x007fffff}}
	for i := 0; i < 2048; i++ {
		pairs = append(pairs, [2]uint32{rng.Uint32(), rng.Uint32()})
	}
	for _, kernel := range []func([]float32, []float32, []float32) error{patterns.XorBitsScalar, patterns.XorBitsSIMD} {
		for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 2057} {
			for alias := -1; alias < 2; alias++ {
				a, b := make([]float32, n), make([]float32, n)
				before := make([][2]uint32, n)
				for i := range a {
					p := pairs[i%len(pairs)]
					before[i] = p
					a[i], b[i] = math.Float32frombits(p[0]), math.Float32frombits(p[1])
				}
				guard := make([]float32, n+2)
				guard[0], guard[n+1] = 123, 456
				out := guard[1 : n+1 : n+1]
				if alias >= 0 {
					out = [][]float32{a, b}[alias]
				}
				if e := kernel(a, b, out); e != nil {
					t.Fatal(e)
				}
				for i, v := range out {
					want := xorTruthBits(before[i][0], before[i][1])
					if math.Float32bits(v) != want {
						t.Fatalf("n=%d alias=%d i=%d got=%08x want=%08x", n, alias, i, math.Float32bits(v), want)
					}
					if alias < 0 && before[i] != ([2]uint32{math.Float32bits(a[i]), math.Float32bits(b[i])}) {
						t.Fatal("input changed")
					}
				}
				if guard[0] != 123 || guard[n+1] != 456 {
					t.Fatal("canary")
				}
			}
		}
		for short := 0; short < 3; short++ {
			arrays := [][]float32{{1}, {2}, {99}}
			arrays[short] = nil
			if e := kernel(arrays[0], arrays[1], arrays[2]); !errors.Is(e, patterns.ErrLengthMismatch) {
				t.Fatal(e)
			}
			if short != 2 && arrays[2][0] != 99 {
				t.Fatal("error wrote output")
			}
		}
		x, out := []float32{math.Float32frombits(0x7f812345)}, make([]float32, 1)
		if count := testing.AllocsPerRun(20, func() {
			if e := kernel(x, x, out); e != nil {
				panic(e)
			}
		}); count != 0 {
			t.Fatal(count)
		}
	}
}
