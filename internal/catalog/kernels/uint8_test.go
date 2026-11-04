package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestSaturatingAdd8TypedBindings(t *testing.T) {
	op, e := Lookup(catalog.SaturatingAdd8, catalog.Uint8)
	if e != nil {
		t.Fatal(e)
	}
	for _, impl := range op.Implementations {
		if impl.BindByteArrays == nil || impl.BindUint8 != nil || impl.Bind != nil || impl.Run != nil {
			t.Fatal("fabricated predicate/float32 input")
		}
		out := []uint8{0}
		if e := impl.BindByteArrays([]uint8{255}, []uint8{1}, out)(); e != nil || out[0] != 255 {
			t.Fatal(e)
		}
		if e := impl.BindByteArrays(nil, []uint8{1}, out)(); e == nil || out[0] != 255 {
			t.Fatal(e)
		}
	}
	if len(uint8Bindings) != 2 {
		t.Fatal("orphan byte binding")
	}
	if _, e := Lookup(catalog.SaturatingAdd8, catalog.Float32); e == nil {
		t.Fatal("implicit float input")
	}
}
