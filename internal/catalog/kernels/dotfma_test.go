package kernels

import (
	"errors"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func TestFusedDotCheckedBindingsAndCounts(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	op, err := Lookup(catalog.DotFMA, catalog.Float32)
	if err != nil || len(op.Implementations) != 4 {
		t.Fatal("fused bindings", err)
	}
	for _, layout := range []int{1, 2} {
		x, y := make([]float32, 8*w), make([]float32, 8*w)
		for i := range y {
			y[i] = 1
		}
		x[0], x[layout*w], x[2*layout*w] = 16777216, 1, -16777216
		for _, impl := range op.Implementations {
			if impl.BindCheckedScalar == nil || impl.BindScalar != nil || impl.Run != nil || impl.BindConfigured != nil {
				t.Fatal("wrong fused binding shape")
			}
			want := float32(0)
			if impl.Descriptor.ID == catalog.SIMDStriped2 && layout == 1 || impl.Descriptor.ID == catalog.SIMDStriped4 && layout == 2 {
				want = 1
			}
			if got, err := impl.BindCheckedScalar(x, y)(); err != nil || got != want {
				t.Fatal("count binding mismatch", impl.Descriptor.ID, got, err)
			}
			if _, err := impl.BindCheckedScalar(x, nil)(); !errors.Is(err, reduction.ErrLengthMismatch) {
				t.Fatal("binding dropped length error")
			}
		}
	}
}
