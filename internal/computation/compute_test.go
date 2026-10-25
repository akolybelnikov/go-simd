package computation_test

import (
	"errors"
	"fmt"
	"simd"
	"slices"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/computation"
)

func TestImplementations(t *testing.T) {
	variants := []struct {
		name string
		run  func([]float32, []float32, []float32) error
	}{
		{"Scalar", computation.ComputeHeavyScalar},
		{"SIMD", computation.ComputeHeavySIMD},
	}
	var vector simd.Float32s
	width := vector.Len()
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			for _, size := range []int{0, 1, width - 1, width, width + 1, 2*width - 1, 2 * width, 2*width + 1, 1025} {
				for _, alias := range []string{"none", "a", "b"} {
					t.Run(fmt.Sprintf("n=%d/alias=%s", size, alias), func(t *testing.T) {
						a, b, out := make([]float32, size), make([]float32, size), make([]float32, size)
						want := make([]float32, size)
						for i := range a {
							x, y := float32(i%11-5), float32(i%7-3)
							a[i], b[i] = x, y
							want[i] = x*x + y*y + x*y
						}
						if alias == "a" {
							out = a
						}
						if alias == "b" {
							out = b
						}
						if err := variant.run(a, b, out); err != nil {
							t.Fatal(err)
						}
						if !slices.Equal(out, want) {
							t.Fatalf("got %v, want %v", out, want)
						}
					})
				}
			}
			t.Run("known values", func(t *testing.T) {
				a, b := []float32{-2, 0, 3}, []float32{4, -5, 2}
				want := []float32{12, 25, 19}
				out := make([]float32, 3)
				if err := variant.run(a, b, out); err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(out, want) {
					t.Fatalf("got %v, want %v", out, want)
				}
			})
			for _, lengths := range [][3]int{{2, 1, 2}, {1, 2, 2}, {2, 2, 1}, {1, 1, 2}} {
				out := make([]float32, lengths[2])
				for i := range out {
					out[i] = 123
				}
				before := slices.Clone(out)
				err := variant.run(make([]float32, lengths[0]), make([]float32, lengths[1]), out)
				if !errors.Is(err, computation.ErrLengthMismatch) {
					t.Fatalf("lengths %v: got %v, want ErrLengthMismatch", lengths, err)
				}
				if !slices.Equal(out, before) {
					t.Fatal("invalid call modified output")
				}
			}
		})
	}
}
