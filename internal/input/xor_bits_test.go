package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"math"
	"testing"
)

func TestXorBitsFixturesAndBothOperands(t *testing.T) {
	op, _ := catalog.ByGroup("xorbits")
	o := DefaultOptions()
	for _, n := range []int{0, 1, 7, 8, 9, 1025} {
		a, b, out := MakeForOperation(op, n, o)
		if len(a) != n || len(b) != n || len(out) != n {
			t.Fatal("shape")
		}
		hash := FingerprintFloat32(a, b)
		if hash != FingerprintForOperation(op, n, o) {
			t.Fatal("stream hash")
		}
		if n > 0 {
			for _, v := range [][]float32{a, b} {
				bits := math.Float32bits(v[0])
				v[0] = math.Float32frombits(bits ^ 1)
				if hash == FingerprintFloat32(a, b) {
					t.Fatal("missing operand bits")
				}
				v[0] = math.Float32frombits(bits)
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
