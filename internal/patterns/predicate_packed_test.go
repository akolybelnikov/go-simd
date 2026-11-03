package patterns_test

import (
	"bytes"
	"errors"
	"math"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func TestPackedPredicateExport(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	values := []float32{0, math.Float32frombits(0x80000000), 1, -1, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), math.Float32frombits(0x7f812345)}
	rng := rand.New(rand.NewPCG(602, 1103))
	for i := 0; i < 100; i++ {
		values = append(values, math.Float32frombits(rng.Uint32()))
	}
	kernels := []func([]float32, []uint8, float32) error{patterns.PredicateGTPackedScalar, patterns.PredicateGTPackedSIMD}
	for _, threshold := range append(values[:12:12], -1.5, 1.5) {
		for _, n := range []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17, 31, 32, 33, 63, 64, 65, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
			x := make([]float32, n)
			before := make([]uint32, n)
			for i := range x {
				x[i] = values[i%len(values)]
				before[i] = math.Float32bits(x[i])
			}
			length := n / 8
			if n%8 != 0 {
				length++
			}
			for _, kernel := range kernels {
				guard := bytes.Repeat([]byte{0xff}, length+2)
				out := guard[1 : length+1 : length+1]
				if err := kernel(x, out, threshold); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < length*8; i++ {
					want := i < n && x[i] > threshold
					if got := out[i/8]&(1<<uint(i%8)) != 0; got != want {
						t.Fatalf("n=%d threshold=%08x bit=%d: %v != %v", n, math.Float32bits(threshold), i, got, want)
					}
				}
				if guard[0] != 0xff || guard[length+1] != 0xff {
					t.Fatal("output canary overwritten")
				}
				for i, v := range x {
					if math.Float32bits(v) != before[i] {
						t.Fatal("input modified")
					}
				}
			}
		}
	}
	for _, kernel := range kernels {
		out := []byte{0xff, 0xff}
		if err := kernel([]float32{1, 0, 0, 1, 0, 0, 1, 0, 1}, out, 0); err != nil || !bytes.Equal(out, []byte{0x49, 0x01}) {
			t.Fatal("LSB/cross-byte order", out, err)
		}
		if err := kernel(make([]float32, 9), out, -1); err != nil || !bytes.Equal(out, []byte{0xff, 0x01}) {
			t.Fatal("true padding", out, err)
		}
		if err := kernel(make([]float32, 9), out, 0); err != nil || !bytes.Equal(out, []byte{0, 0}) {
			t.Fatal("stale output not overwritten", out, err)
		}
		if err := kernel([]float32{-2}, out[:1], -1); err != nil || out[0] != 0 {
			t.Fatal("padded zero input escaped", out, err)
		}
		if err := kernel(nil, nil, 0); err != nil {
			t.Fatal(err)
		}
		for _, lengths := range [][2]int{{0, 1}, {1, 0}, {8, 2}, {9, 1}, {9, 3}} {
			out := bytes.Repeat([]byte{0xa5}, lengths[1])
			before := bytes.Clone(out)
			if err := kernel(make([]float32, lengths[0]), out, 0); !errors.Is(err, patterns.ErrLengthMismatch) || !bytes.Equal(out, before) {
				t.Fatal("length validation wrote output", err)
			}
		}
		x, out := make([]float32, 65), make([]byte, 9)
		if w <= 16 && testing.AllocsPerRun(20, func() {
			if err := kernel(x, out, -1); err != nil {
				panic(err)
			}
		}) != 0 {
			t.Fatal("unexpected allocation")
		}
	}
}
