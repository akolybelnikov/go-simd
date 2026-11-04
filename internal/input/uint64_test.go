package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"

	"testing"
)

func TestClmul64TrueFixturesAndFullFingerprint(t *testing.T) {
	op, _ := catalog.ByGroup("clmul64")
	o := DefaultOptions()
	for _, n := range []int{0, 1, 3, 4, 5, 1025} {
		a, b, out := MakeClmul64(n)
		if len(a) != n || len(b) != n || len(out) != n {
			t.Fatal("shape")
		}
		hash := FingerprintUint64(a, b)
		if hash != FingerprintForOperation(op, n, o) {
			t.Fatal("stream hash")
		}
		if n > 0 {
			for _, v := range [][]uint64{a, b} {
				bits := v[0]
				v[0] = bits ^ 1
				if hash == FingerprintUint64(a, b) {
					t.Fatal("lost uint64 input bits")
				}
				v[0] = bits
			}
		}
	}
	before := FingerprintForOperation(op, 1025, o) + o.Generator(op)
	o.Alpha++
	o.Pattern = "shuffled"
	o.Seed++
	o.FMAInput = "rounding"
	o.SumInput = "cancellation"
	if before != FingerprintForOperation(op, 1025, o)+o.Generator(op) {
		t.Fatal("irrelevant flags")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("float32 helper accepted uint64")
		}
	}()
	MakeForOperation(op, 1, o)
}
