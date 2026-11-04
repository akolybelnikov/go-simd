package kernels

import (
	"errors"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"testing"
)

func TestTypedUint64Bindings(t *testing.T) {
	op, e := Lookup(catalog.Clmul64, catalog.Uint64)
	if e != nil {
		t.Fatal(e)
	}
	for _, impl := range op.Implementations {
		if impl.BindUint64 == nil || impl.Bind != nil || impl.Run != nil {
			t.Fatal("typed shape")
		}
		out := []uint64{99}
		if e := impl.BindUint64([]uint64{1}, []uint64{2}, out)(); e != nil || out[0] != 2 {
			t.Fatal(e)
		}
		if e := impl.BindUint64(nil, []uint64{2}, out)(); !errors.Is(e, integer.ErrLengthMismatch) || out[0] != 2 {
			t.Fatal(e)
		}
	}
	if len(clmul64Bindings) != 2 {
		t.Fatal("orphan binding")
	}
	if _, e := Lookup(catalog.Clmul64, catalog.Float32); e == nil {
		t.Fatal("implicit type cast")
	}
}
