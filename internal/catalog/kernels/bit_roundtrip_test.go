package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestBitRoundTripUnaryBinding(t *testing.T) {
	op, e := Lookup(catalog.BitRoundTrip, catalog.Float32)
	if e != nil {
		t.Fatal(e)
	}
	x, _, out := input.MakeForOperation(op.Descriptor, 17, input.DefaultOptions())
	want := input.FingerprintFloat32(x, nil)
	for _, impl := range op.Implementations {
		if impl.BindUnary == nil || impl.Bind != nil || impl.Run != nil || impl.BindConfigured != nil {
			t.Fatal("binding shape")
		}
		if e := impl.BindUnary(x, out)(); e != nil || input.FingerprintFloat32(out, nil) != want {
			t.Fatal(e)
		}
	}
	if len(bitRoundTripBindings) != 2 {
		t.Fatal("orphan binding")
	}
}
