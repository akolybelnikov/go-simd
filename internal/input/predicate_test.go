package input

import (
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestPredicateByteFixtureIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("predicate")
	o := DefaultOptions()
	o.Pattern = "shuffled"
	o.Seed = 123
	o.Threshold = math.Float32bits(-1.5)
	x, y, out := MakeForOperation(op, 1025, o)
	if len(x) != 1025 || y != nil || out != nil || FingerprintFloat32(x, nil) != FingerprintForOperation(op, 1025, o) {
		t.Fatal("typed shape/fingerprint")
	}
	base := o.Identity(op) + o.Generator(op) + FingerprintForOperation(op, 1025, o)
	o.Alpha++
	o.Beta++
	o.Lower++
	o.Upper++
	o.DotInput = "fusion-sensitive"
	o.SumInput = "cancellation"
	o.FMAInput = "rounding"
	if base != o.Identity(op)+o.Generator(op)+FingerprintForOperation(op, 1025, o) {
		t.Fatal("irrelevant settings changed input")
	}
	for _, bits := range []uint32{0x7f800000, 0x7fc12345} {
		o.Threshold = bits
		o.Pattern = "all"
		if o.ValidateFor(op) == nil {
			t.Fatal("impossible true fixture accepted")
		}
		o.Pattern = "none"
		if err := o.ValidateFor(op); err != nil {
			t.Fatal(err)
		}
	}
}
