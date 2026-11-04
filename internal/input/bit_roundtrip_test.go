package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestBitRoundTripFixtureAndIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("bitroundtrip")
	o := DefaultOptions()
	for _, n := range []int{0, 1, 7, 8, 9, 1025} {
		x, y, out := MakeForOperation(op, n, o)
		if len(x) != n || y != nil || len(out) != n || FingerprintFloat32(x, nil) != FingerprintForOperation(op, n, o) {
			t.Fatal("arity/hash")
		}
	}
	before := o.Generator(op) + FingerprintForOperation(op, 1025, o)
	o.Alpha++
	o.FMAInput = "rounding"
	o.Seed++
	o.Pattern = "shuffled"
	if before != o.Generator(op)+FingerprintForOperation(op, 1025, o) {
		t.Fatal("irrelevant options")
	}
	if o.Generator(op) == CopyGeneratorID {
		t.Fatal("copy and roundtrip generator namespace pooled")
	}
}
