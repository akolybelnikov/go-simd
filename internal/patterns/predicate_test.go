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

func TestPredicateExport(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	values := []float32{0, math.Float32frombits(0x80000000), 1, -1, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, math.MaxFloat32, -math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc12345), math.Float32frombits(0x7f812345)}
	rng := rand.New(rand.NewPCG(501, 1103))
	for i := 0; i < 100; i++ {
		values = append(values, math.Float32frombits(rng.Uint32()))
	}
	for _, threshold := range append(values[:12:12], -1.5, 1.5) {
		for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
			x := make([]float32, n)
			before := make([]uint32, n)
			want := make([]byte, n)
			for i := range x {
				x[i] = values[i%len(values)]
				before[i] = math.Float32bits(x[i])
				if x[i] > threshold {
					want[i] = 1
				}
			}
			for _, kernel := range []func([]float32, []uint8, float32) error{patterns.PredicateGTScalar, patterns.PredicateGTSIMD} {
				guard := bytes.Repeat([]byte{0xa5}, n+2)
				out := guard[1 : n+1 : n+1]
				if err := kernel(x, out, threshold); err != nil || !bytes.Equal(out, want) {
					t.Fatalf("n=%d t=%08x: %v got=%v want=%v", n, math.Float32bits(threshold), err, out, want)
				}
				if guard[0] != 0xa5 || guard[n+1] != 0xa5 {
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
	for _, kernel := range []func([]float32, []uint8, float32) error{patterns.PredicateGTScalar, patterns.PredicateGTSIMD} {
		if err := kernel(nil, nil, 0); err != nil {
			t.Fatal(err)
		}
		for _, lengths := range [][2]int{{0, 1}, {1, 0}, {3, 2}, {2, 3}} {
			out := bytes.Repeat([]byte{0xa5}, lengths[1])
			before := bytes.Clone(out)
			if err := kernel(make([]float32, lengths[0]), out, 0); !errors.Is(err, patterns.ErrLengthMismatch) || !bytes.Equal(out, before) {
				t.Fatal("length failure wrote output", err)
			}
		}
		x, out := make([]float32, 2*w+1), make([]uint8, 2*w+1)
		if w <= 16 && testing.AllocsPerRun(20, func() {
			if err := kernel(x, out, -1); err != nil {
				panic(err)
			}
		}) != 0 {
			t.Fatal("unexpected allocation")
		}
		// Padded zero lanes compare true at -1: only the one valid byte may escape.
		guard := []byte{7, 7, 7}
		if err := kernel([]float32{-2}, guard[1:2:2], -1); err != nil || !bytes.Equal(guard, []byte{7, 0, 7}) {
			t.Fatal("inactive lane escaped", err, guard)
		}
	}
}
