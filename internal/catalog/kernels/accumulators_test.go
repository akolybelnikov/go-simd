package kernels

import (
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestAccumulatorBindingsMatchDeclaredCounts(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	for _, operation := range []string{catalog.Sum, catalog.Dot} {
		op, err := Lookup(operation, catalog.Float32)
		if err != nil {
			t.Fatal(err)
		}
		for _, layout := range []int{1, 2} {
			x, y := make([]float32, 8*w), make([]float32, 8*w)
			for i := range y {
				y[i] = 1
			}
			x[0], x[layout*w], x[2*layout*w] = 16777216, 1, -16777216
			for _, impl := range op.Implementations {
				want := float32(0)
				if impl.Descriptor.ID == catalog.SIMDStriped2 && layout == 1 || impl.Descriptor.ID == catalog.SIMDStriped4 && layout == 2 {
					want = 1
				}
				var got float32
				if operation == catalog.Sum {
					got = impl.BindScalar(x)()
				} else {
					got, err = impl.BindCheckedScalar(x, y)()
					if err != nil {
						t.Fatal(err)
					}
				}
				if got != want {
					t.Fatalf("wrong accumulator binding %s/%s/layout%d: %g != %g", operation, impl.Descriptor.ID, layout, got, want)
				}
			}
		}
	}
}
