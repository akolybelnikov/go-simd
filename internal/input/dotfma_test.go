package input

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestFusedDotSharedFixturesAndIdentity(t *testing.T) {
	fused, _ := catalog.ByGroup("dotfma")
	unfused, _ := catalog.ByGroup("dot")
	for _, fixture := range []string{"balanced", "cancellation", "mixed-magnitude", "fusion-sensitive", ""} {
		o := DefaultOptions()
		o.DotInput = fixture
		for _, n := range []int{0, 1, 64, 1025} {
			x, y, out := MakeForOperation(fused, n, o)
			a, b, _ := MakeForOperation(unfused, n, o)
			if len(x) != n || len(y) != n || out != nil || FingerprintFloat32(x, y) != FingerprintForOperation(fused, n, o) || FingerprintFloat32(x, y) != FingerprintFloat32(a, b) {
				t.Fatal("fused/unfused fixtures differ")
			}
			if o.Generator(fused) == o.Generator(unfused) {
				t.Fatal("operation fixture identity conflated")
			}
			changed := o
			changed.FMAInput = "rounding"
			changed.Alpha = 0
			changed.Seed = 123
			if changed.Generator(fused) != o.Generator(fused) || FingerprintForOperation(fused, n, changed) != FingerprintForOperation(fused, n, o) {
				t.Fatal("irrelevant flags change fused dot")
			}
		}
	}
}
