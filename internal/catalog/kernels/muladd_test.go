package kernels

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
	"testing"
)

func TestMulAddTypedBindings(t *testing.T) {
	op, e := Lookup(catalog.MulAdd, catalog.Float32)
	if e != nil {
		t.Fatal(e)
	}
	a, b, c, out := []float32{1 + 0x1p-23}, []float32{1 - 0x1p-23}, []float32{-1}, []float32{99}
	for _, impl := range op.Implementations {
		if impl.BindTernary == nil || impl.Bind != nil || impl.BindConfigured != nil || impl.Run != nil {
			t.Fatal("binding shape")
		}
		if e := impl.BindTernary(a, b, c, out)(); e != nil || math.Float32bits(out[0]) != math.Float32bits(fma32.FMA(a[0], b[0], c[0])) {
			t.Fatal(e)
		}
		out[0] = 99
		if e := impl.BindTernary(a, b, nil, out)(); !errors.Is(e, patterns.ErrLengthMismatch) || out[0] != 99 {
			t.Fatal(e)
		}
	}
	if len(ternaryBindings) != 2 {
		t.Fatal("orphan ternary binding")
	}
}
