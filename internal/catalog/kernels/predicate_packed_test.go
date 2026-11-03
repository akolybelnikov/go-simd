package kernels

import (
	"bytes"
	"errors"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func TestPackedPredicateBindings(t *testing.T) {
	op, err := Lookup(catalog.PredicateGTPacked, catalog.Float32)
	if err != nil || len(op.Implementations) != 2 {
		t.Fatal(err)
	}
	for _, impl := range op.Implementations {
		if impl.BindUint8 == nil || impl.BindConfigured != nil || impl.Run != nil || impl.BindScalar != nil || impl.BindCheckedScalar != nil {
			t.Fatal("wrong typed packed binding")
		}
		out := []byte{255, 255}
		if err := impl.BindUint8([]float32{1, 0, 0, 1, 0, 0, 1, 0, 1}, out, input.DefaultOptions())(); err != nil || !bytes.Equal(out, []byte{0x49, 1}) {
			t.Fatal(out, err)
		}
		if err := impl.BindUint8(nil, out, input.DefaultOptions())(); !errors.Is(err, patterns.ErrLengthMismatch) {
			t.Fatal(err)
		}
	}
}
