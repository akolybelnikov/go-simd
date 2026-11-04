package kernels

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/movement"
	"math"
	"testing"
)

func TestUnaryFloat64Binding(t *testing.T) {
	op, e := Lookup(catalog.Copy64, catalog.Float64)
	if e != nil {
		t.Fatal(e)
	}
	if len(copy64Bindings) != 3 {
		t.Fatal("orphan")
	}
	for _, impl := range op.Implementations {
		if impl.BindUnaryFloat64 == nil || impl.BindFloat64 != nil || impl.Run != nil {
			t.Fatal("unary type")
		}
		x := []float64{math.Float64frombits(0x7ff0123456789abc)}
		out := []float64{99}
		if e := impl.BindUnaryFloat64(x, out)(); e != nil || math.Float64bits(out[0]) != math.Float64bits(x[0]) {
			t.Fatal(e)
		}
		if e := impl.BindUnaryFloat64(nil, out)(); !errors.Is(e, movement.ErrLengthMismatch) || math.Float64bits(out[0]) != math.Float64bits(x[0]) {
			t.Fatal(e)
		}
	}
	if _, e := Lookup(catalog.Copy64, catalog.Float32); e == nil {
		t.Fatal("implicit cast")
	}
}
