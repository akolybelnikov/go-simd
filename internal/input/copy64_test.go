package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"math"
	"testing"
)

func TestCopy64FullUnaryBits(t *testing.T) {
	op, _ := catalog.ByGroup("copy64")
	opts := DefaultOptions()
	for _, n := range []int{0, 1, 3, 4, 5, 13, 1025} {
		x, out := MakeCopy64(n)
		if len(x) != n || len(out) != n || FingerprintFloat64(x, nil) != FingerprintForOperation(op, n, opts) {
			t.Fatal("extent/hash")
		}
		if n > 0 {
			before := FingerprintFloat64(x, nil)
			x[0] = math.Float64frombits(math.Float64bits(x[0]) ^ 1)
			if before == FingerprintFloat64(x, nil) {
				t.Fatal("lost bit")
			}
		}
	}
	before := FingerprintForOperation(op, 1025, opts) + opts.Generator(op)
	opts.Alpha++
	opts.Pattern = "shuffled"
	opts.Seed++
	opts.SumInput = "cancellation"
	if before != FingerprintForOperation(op, 1025, opts)+opts.Generator(op) {
		t.Fatal("irrelevant flags")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("float32 helper accepted binary64")
		}
	}()
	MakeForOperation(op, 1, opts)
}
