package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"math"
	"testing"
)

func TestXorBitsTypedRunAndBinding(t *testing.T) {
	op, e := Lookup(catalog.XorBits, catalog.Float32)
	if e != nil {
		t.Fatal(e)
	}
	a, b, out := []float32{1}, []float32{math.Float32frombits(0x80000000)}, []float32{0}
	for _, impl := range op.Implementations {
		if impl.Run == nil || impl.Bind == nil || impl.BindConfigured != nil {
			t.Fatal("binary binding")
		}
		for _, run := range []func() error{func() error { return impl.Run(a, b, out) }, impl.Bind(a, b, out)} {
			if e := run(); e != nil || math.Float32bits(out[0]) != 0xbf800000 {
				t.Fatal(e)
			}
		}
	}
	if len(xorBitsBindings) != 2 {
		t.Fatal("orphan XOR binding")
	}
}
