package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestSum8UnaryProjectedBinding(t *testing.T) {
	op, e := Lookup(catalog.Sum8, catalog.Float32)
	if e != nil {
		t.Fatal(e)
	}
	x := []float32{-1e20, 1, 1e20, 0, 1e20, 0, -1e20, 0, 2}
	for _, impl := range op.Implementations {
		if impl.BindUnary == nil || impl.BindScalar != nil || impl.Run != nil {
			t.Fatal("projected array binding")
		}
		out := make([]float32, 2)
		if e := impl.BindUnary(x, out)(); e != nil || out[0] != 1 || out[1] != 2 {
			t.Fatal(out, e)
		}
	}
	if len(sum8Bindings) != 2 {
		t.Fatal("orphan binding")
	}
}
