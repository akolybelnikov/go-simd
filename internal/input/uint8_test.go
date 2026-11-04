package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestSaturatingAdd8TrueBytesAndHashes(t *testing.T) {
	op, _ := catalog.ByGroup("saturatingadd8")
	o := DefaultOptions()
	for _, n := range []int{0, 1, 15, 16, 17, 32, 33, 1025} {
		a, b, out := MakeSaturatingAdd8(n)
		if len(a) != n || len(b) != n || len(out) != n {
			t.Fatal("shape")
		}
		hash := FingerprintUint8(a, b)
		if hash != FingerprintForOperation(op, n, o) {
			t.Fatal("stream hash")
		}
		if n > 0 {
			for _, operand := range [][]uint8{a, b} {
				operand[0] ^= 1
				if hash == FingerprintUint8(a, b) {
					t.Fatal("missing operand")
				}
				operand[0] ^= 1
			}
		}
	}
	before := o.Generator(op) + FingerprintForOperation(op, 1025, o)
	o.Alpha++
	o.Pattern = "shuffled"
	o.Seed++
	o.FMAInput = "rounding"
	if before != o.Generator(op)+FingerprintForOperation(op, 1025, o) {
		t.Fatal("irrelevant options")
	}
}
