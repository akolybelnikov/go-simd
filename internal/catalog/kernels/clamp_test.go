package kernels

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
	"testing"
)

func TestClampConfiguredBindings(t *testing.T) {
	op, err := Lookup(catalog.Clamp, catalog.Float32)
	if err != nil {
		t.Fatal(err)
	}
	options := input.DefaultOptions()
	options.Lower = 0
	options.Upper = 0x80000000
	x := []float32{-1, 1, math.Float32frombits(0x80000000), 0, math.Float32frombits(0x7f812345)}
	want := []uint32{0, 0x80000000, 0x80000000, 0, 0x7f812345}
	for _, impl := range op.Implementations {
		out := make([]float32, len(x))
		run := impl.BindConfigured(x, nil, out, options)
		if err := run(); err != nil {
			t.Fatal(err)
		}
		for i, v := range out {
			if math.Float32bits(v) != want[i] {
				t.Fatal(impl.Descriptor.Label, i)
			}
		}
		if err := impl.BindConfigured(x, nil, out[:1], options)(); !errors.Is(err, patterns.ErrLengthMismatch) {
			t.Fatal(err)
		}
		bad := options
		bad.Lower = math.Float32bits(1)
		if err := impl.BindConfigured(x, nil, out, bad)(); !errors.Is(err, patterns.ErrInvalidBounds) {
			t.Fatal(err)
		}
	}
}
