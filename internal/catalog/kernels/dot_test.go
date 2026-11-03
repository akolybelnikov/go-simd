package kernels

import (
	"errors"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func TestDotCheckedBindings(t *testing.T) {
	op, err := Lookup(catalog.Dot, catalog.Float32)
	if err != nil || len(op.Implementations) != 4 {
		t.Fatal("dot binding inventory", err)
	}
	for _, impl := range op.Implementations {
		if impl.BindCheckedScalar == nil || impl.BindScalar != nil || impl.Bind != nil || impl.Run != nil || impl.BindConfigured != nil {
			t.Fatal("dot binding shape")
		}
		if got, err := impl.BindCheckedScalar([]float32{1, 2, 3}, []float32{4, 5, 6})(); err != nil || got != 32 {
			t.Fatal("dot binding result", err, got)
		}
		if got, err := impl.BindCheckedScalar([]float32{1}, nil)(); !errors.Is(err, reduction.ErrLengthMismatch) || got != 0 {
			t.Fatal("bound length error lost", err)
		}
	}
}
