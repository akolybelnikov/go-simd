package kernels

import (
	"errors"
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/movement"
)

func TestUnaryCopyBindings(t *testing.T) {
	op, err := Lookup(catalog.Copy, catalog.Float32)
	if err != nil || len(op.Implementations) != 4 {
		t.Fatal(err)
	}
	x := []float32{math.Float32frombits(0x80000000), math.Float32frombits(0x7f812345), 1}
	for _, impl := range op.Implementations {
		if impl.BindUnary == nil || impl.BindConfigured != nil || impl.Run != nil || impl.Bind != nil || impl.BindUint8 != nil || impl.BindScalar != nil || impl.BindCheckedScalar != nil {
			t.Fatal("incorrect copy binding")
		}
		out := make([]float32, len(x))
		if err := impl.BindUnary(x, out)(); err != nil {
			t.Fatal(err)
		}
		for i, v := range out {
			if math.Float32bits(v) != math.Float32bits(x[i]) {
				t.Fatal("copy bits changed")
			}
		}
		if err := impl.BindUnary(nil, out)(); !errors.Is(err, movement.ErrLengthMismatch) {
			t.Fatal(err)
		}
	}
}
