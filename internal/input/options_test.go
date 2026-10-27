package input

import (
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestPatternFixturesAndFingerprints(t *testing.T) {
	for _, op := range catalog.All() {
		for _, pattern := range []string{"none", "all", "alternating", "clustered", "shuffled"} {
			for _, n := range []int{0, 1, 7, 8, 9, 1025} {
				o := DefaultOptions()
				o.Pattern = pattern
				if err := o.ValidateFor(op); err != nil {
					t.Fatal(err)
				}
				a, b, out := MakeForOperation(op, n, o)
				if len(a) != n || len(out) != n || len(op.Inputs) == 2 && len(b) != n || len(op.Inputs) == 1 && len(b) != 0 {
					t.Fatal("operand arity mismatch")
				}
				if got, want := FingerprintFloat32(a, b), FingerprintForOperation(op, n, o); got != want {
					t.Fatalf("fingerprint drift: %s/%s/n=%d", op.ID, pattern, n)
				}
				if IsPredicate(op) {
					count := 0
					for _, v := range a {
						if o.matches(op, v) {
							count++
						}
					}
					switch pattern {
					case "none":
						if count != 0 {
							t.Fatal("none fixture has true lanes")
						}
					case "all":
						if count != n {
							t.Fatal("all fixture has false lanes")
						}
					case "alternating", "clustered":
						if count != (n+1)/2 {
							t.Fatal("balanced pattern drift")
						}
					case "shuffled":
						if n == 1025 && (count < 400 || count > 625) {
							t.Fatal("shuffled distribution drift")
						}
					}
				}
			}
		}
	}
}

func TestParameterBitsAndImpossiblePredicate(t *testing.T) {
	op, _ := catalog.Lookup(catalog.Scale, catalog.Float32)
	positive, negative := DefaultOptions(), DefaultOptions()
	positive.Alpha = 0
	negative.Alpha = 0x80000000
	if positive.Identity(op) == negative.Identity(op) {
		t.Fatal("signed-zero identity collapsed")
	}
	positive.Alpha = 0x7fc12345
	negative.Alpha = 0x7fc54321
	if positive.Identity(op) == negative.Identity(op) {
		t.Fatal("NaN-payload identity collapsed")
	}
	threshold, _ := catalog.Lookup(catalog.ThresholdZero, catalog.Float32)
	for _, bits := range []uint32{math.Float32bits(float32(math.Inf(1))), 0x7fc12345} {
		o := DefaultOptions()
		o.Threshold = bits
		if err := o.ValidateFor(threshold); err == nil {
			t.Fatal("impossible true fixture accepted")
		}
		o.Pattern = "none"
		if err := o.ValidateFor(threshold); err != nil {
			t.Fatal(err)
		}
	}
	rangeOp, _ := catalog.Lookup(catalog.RangeZero, catalog.Float32)
	o := DefaultOptions()
	o.Lower = o.Upper
	if err := o.ValidateFor(rangeOp); err == nil {
		t.Fatal("empty true range accepted")
	}
	o.Pattern = "none"
	if err := o.ValidateFor(rangeOp); err != nil {
		t.Fatal(err)
	}
	o.Pattern = "unknown"
	if err := o.Validate(); err == nil {
		t.Fatal("unknown pattern accepted")
	}
}
