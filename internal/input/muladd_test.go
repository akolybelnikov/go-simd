package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"math"
	"testing"
)

func TestMulAddFixturesAllOperandsAndIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("muladd")
	for _, fixture := range []string{"balanced", "rounding", "mixed-magnitude", ""} {
		o := DefaultOptions()
		o.FMAInput = fixture
		for _, n := range []int{0, 1, 7, 8, 9, 64, 1025} {
			a, b, c, out := MakeMulAdd(n, o)
			if len(a) != n || len(b) != n || len(c) != n || len(out) != n {
				t.Fatal("arity")
			}
			hash := FingerprintTernaryFloat32(a, b, c)
			if hash != FingerprintForOperation(op, n, o) {
				t.Fatal("streamed fingerprint drift")
			}
			if n > 0 {
				for _, operand := range [][]float32{a, b, c} {
					bits := math.Float32bits(operand[0])
					operand[0] = math.Float32frombits(bits ^ 1)
					if hash == FingerprintTernaryFloat32(a, b, c) {
						t.Fatal("operand missing from fingerprint")
					}
					operand[0] = math.Float32frombits(bits)
				}
			}
			before := o.Generator(op) + FingerprintForOperation(op, n, o)
			irrelevant := o
			irrelevant.Alpha++
			irrelevant.Beta++
			irrelevant.Threshold++
			irrelevant.Pattern = "shuffled"
			irrelevant.Seed++
			irrelevant.SumInput = "cancellation"
			irrelevant.DotInput = "fusion-sensitive"
			if before != irrelevant.Generator(op)+FingerprintForOperation(op, n, irrelevant) || o.Identity(op) != "" {
				t.Fatal("irrelevant settings change ternary identity")
			}
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("two-operand helper silently dropped third input")
		}
	}()
	MakeForOperation(op, 1, DefaultOptions())
}
