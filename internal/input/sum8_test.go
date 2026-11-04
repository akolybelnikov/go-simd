package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func TestSum8FixtureProjectedOutputAndIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("sum8")
	for _, fixture := range []string{"balanced", "cancellation", "mixed-magnitude"} {
		o := DefaultOptions()
		o.SumInput = fixture
		for _, n := range []int{0, 1, 7, 8, 9, 15, 16, 17, 1025} {
			x, y, out := MakeForOperation(op, n, o)
			count, _ := op.OutputElements(n)
			if len(x) != n || y != nil || len(out) != count || FingerprintFloat32(x, nil) != FingerprintForOperation(op, n, o) {
				t.Fatal("shape/hash")
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
}
