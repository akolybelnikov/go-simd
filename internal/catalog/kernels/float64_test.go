package kernels

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestTypedFloat64Bindings(t *testing.T) {
	op, e := Lookup(catalog.Add64, catalog.Float64)
	if e != nil {
		t.Fatal(e)
	}
	for _, impl := range op.Implementations {
		if impl.BindFloat64 == nil || impl.Bind != nil || impl.Run != nil {
			t.Fatal("typed shape")
		}
		out := []float64{99}
		if e := impl.BindFloat64([]float64{1}, []float64{2}, out)(); e != nil || out[0] != 3 {
			t.Fatal(e)
		}
		if e := impl.BindFloat64(nil, []float64{2}, out)(); !errors.Is(e, addition.ErrLengthMismatch) || out[0] != 3 {
			t.Fatal(e)
		}
	}
	if len(float64Bindings) != 2 {
		t.Fatal("orphan binding")
	}
	if _, e := Lookup(catalog.Add64, catalog.Float32); e == nil {
		t.Fatal("implicit type cast")
	}
}
