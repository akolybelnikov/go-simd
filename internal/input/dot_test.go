package input

import (
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestDotFixturesAndRelevantIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("dot")
	seen := map[string]bool{}
	for _, fixture := range []string{"balanced", "cancellation", "mixed-magnitude", ""} {
		for _, n := range []int{0, 1, 64, 1025} {
			o := DefaultOptions()
			o.DotInput = fixture
			if err := o.ValidateFor(op); err != nil {
				t.Fatal(err)
			}
			x, y, out := MakeForOperation(op, n, o)
			if len(x) != n || len(y) != n || out != nil {
				t.Fatal("dot arity/output mismatch")
			}
			hash := FingerprintFloat32(x, y)
			if hash != FingerprintForOperation(op, n, o) {
				t.Fatal("dot streamed fingerprint drift")
			}
			changed := o
			changed.SumInput = "mixed-magnitude"
			changed.Seed = 999
			changed.Alpha = 0
			if FingerprintForOperation(op, n, changed) != hash || changed.Generator(op) != o.Generator(op) {
				t.Fatal("irrelevant settings change dot identity")
			}
			if n == 64 && fixture != "" {
				if seen[hash] {
					t.Fatal("distinct dot fixtures collapsed")
				}
				seen[hash] = true
				y[0] = math.Float32frombits(math.Float32bits(y[0]) ^ 1)
				if FingerprintFloat32(x, y) == hash {
					t.Fatal("dot y missing from fingerprint")
				}
			}
		}
	}
	o := DefaultOptions()
	o.DotInput = "unknown"
	if err := o.Validate(); err == nil {
		t.Fatal("invalid dot fixture accepted")
	}
}
