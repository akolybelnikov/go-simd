package input

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestFMAFixtureIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("affinefma")
	seen := map[string]bool{}
	for _, fixture := range []string{"balanced", "rounding", "mixed-magnitude", ""} {
		for _, n := range []int{0, 1, 8, 65} {
			o := DefaultOptions()
			o.FMAInput = fixture
			x, y, out := MakeForOperation(op, n, o)
			if len(x) != n || len(out) != n || y != nil || FingerprintFloat32(x, nil) != FingerprintForOperation(op, n, o) {
				t.Fatal("FMA fixture shape/hash")
			}
			if n == 65 && fixture != "" {
				hash := FingerprintFloat32(x, nil)
				if seen[hash] {
					t.Fatal("FMA fixtures collapsed")
				}
				seen[hash] = true
			}
			changed := o
			changed.SumInput = "mixed-magnitude"
			changed.Seed = 123
			if changed.Generator(op) != o.Generator(op) || FingerprintForOperation(op, n, changed) != FingerprintForOperation(op, n, o) {
				t.Fatal("irrelevant FMA controls change identity")
			}
		}
	}
	o := DefaultOptions()
	o.FMAInput = "unknown"
	if err := o.Validate(); err == nil {
		t.Fatal("unknown FMA fixture")
	}
}
