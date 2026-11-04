package integer_test

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"simd"
	"slices"
	"testing"
)

func TestSaturatingAdd8EveryPairAndBoundaries(t *testing.T) {
	a, b, want := make([]uint8, 65536), make([]uint8, 65536), make([]uint8, 65536)
	for x := 0; x < 256; x++ {
		for y := 0; y < 256; y++ {
			i := x*256 + y
			a[i], b[i] = uint8(x), uint8(y)
			want[i] = uint8(min(x+y, 255))
		}
	}
	var vector simd.Uint8s
	w := vector.Len()
	for _, kernel := range []func([]uint8, []uint8, []uint8) error{integer.SaturatingAdd8Scalar, integer.SaturatingAdd8SIMD} {
		out := make([]uint8, len(a))
		if e := kernel(a, b, out); e != nil || !slices.Equal(out, want) {
			t.Fatal("exhaustive saturation mismatch", e)
		}
		for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
			for alias := -1; alias < 2; alias++ {
				x, y := slices.Clone(a[127*256:127*256+n]), slices.Clone(b[127*256:127*256+n])
				beforeX, beforeY := slices.Clone(x), slices.Clone(y)
				guard := make([]uint8, n+2)
				guard[0], guard[n+1] = 17, 39
				out := guard[1 : n+1 : n+1]
				if alias >= 0 {
					out = [][]uint8{x, y}[alias]
				}
				if e := kernel(x, y, out); e != nil {
					t.Fatal(e)
				}
				for i, v := range out {
					expected := uint8(min(int(beforeX[i])+int(beforeY[i]), 255))
					if v != expected {
						t.Fatalf("n=%d alias=%d i=%d got=%d want=%d", n, alias, i, v, expected)
					}
				}
				if guard[0] != 17 || guard[n+1] != 39 {
					t.Fatal("canary")
				}
				if alias < 0 && (!slices.Equal(x, beforeX) || !slices.Equal(y, beforeY)) {
					t.Fatal("input changed")
				}
			}
		}
		for short := 0; short < 3; short++ {
			arrays := [][]uint8{{1}, {2}, {99}}
			arrays[short] = nil
			if e := kernel(arrays[0], arrays[1], arrays[2]); !errors.Is(e, integer.ErrLengthMismatch) {
				t.Fatal(e)
			}
			if short != 2 && arrays[2][0] != 99 {
				t.Fatal("error wrote output")
			}
		}
		x, out := []uint8{0, 128, 255}, make([]uint8, 3)
		if count := testing.AllocsPerRun(20, func() {
			if e := kernel(x, x, out); e != nil {
				panic(e)
			}
		}); count != 0 {
			t.Fatal(count)
		}
	}
}
