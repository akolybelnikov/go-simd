package kernels

import (
	"bytes"
	"errors"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func TestPredicateBindings(t *testing.T) {
	op, err := Lookup(catalog.PredicateGT, catalog.Float32)
	if err != nil || len(op.Implementations) != 2 {
		t.Fatal(err)
	}
	o := input.DefaultOptions()
	for _, impl := range op.Implementations {
		if impl.BindUint8 == nil || impl.BindConfigured != nil || impl.Run != nil || impl.Bind != nil || impl.BindScalar != nil || impl.BindCheckedScalar != nil {
			t.Fatal("wrong typed binding")
		}
		out := make([]byte, 4)
		if err := impl.BindUint8([]float32{-1, 0, 1, 2}, out, o)(); err != nil || !bytes.Equal(out, []byte{0, 0, 1, 1}) {
			t.Fatal(out, err)
		}
		if err := impl.BindUint8(nil, []byte{7}, o)(); !errors.Is(err, patterns.ErrLengthMismatch) {
			t.Fatal(err)
		}
	}
}
