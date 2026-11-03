package input

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestPackedPredicateSharesDeclaredInput(t *testing.T) {
	packed, _ := catalog.ByGroup("predicatepack")
	plain, _ := catalog.ByGroup("predicate")
	for _, pattern := range []string{"none", "all", "alternating", "clustered", "shuffled"} {
		o := DefaultOptions()
		o.Pattern = pattern
		o.Seed = 123
		o.Threshold = 0xbfc00000
		x, y, out := MakeForOperation(packed, 65, o)
		if len(x) != 65 || y != nil || out != nil || FingerprintFloat32(x, nil) != FingerprintForOperation(packed, 65, o) || FingerprintForOperation(plain, 65, o) != FingerprintForOperation(packed, 65, o) || o.Generator(plain) != o.Generator(packed) || o.Identity(plain) != o.Identity(packed) {
			t.Fatal("packed fixture differs from byte mask")
		}
	}
}
