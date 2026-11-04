package input

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"math"
	"testing"
)

func TestAddition64TrueFixturesAndFullFingerprint(t *testing.T) {
	op, _ := catalog.ByGroup("addition64")
	o := DefaultOptions()
	for _, n := range []int{0, 1, 3, 4, 5, 1025} {
		a, b, out := MakeAddition64(n)
		if len(a) != n || len(b) != n || len(out) != n {
			t.Fatal("shape")
		}
		hash := FingerprintFloat64(a, b)
		if hash != FingerprintForOperation(op, n, o) {
			t.Fatal("stream hash")
		}
		if n > 0 {
			for _, v := range [][]float64{a, b} {
				bits := math.Float64bits(v[0])
				v[0] = math.Float64frombits(bits ^ 1)
				if hash == FingerprintFloat64(a, b) {
					t.Fatal("lost binary64 input bits")
				}
				v[0] = math.Float64frombits(bits)
			}
		}
	}
	a, _, _ := MakeAddition64(2)
	if a[1] == float64(float32(a[1])) {
		t.Fatal("fixture silently widened float32")
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
			t.Fatal("float32 helper accepted float64")
		}
	}()
	MakeForOperation(op, 1, o)
}
