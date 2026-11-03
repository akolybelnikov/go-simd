package input

import (
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestClampFixturesAndIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("clamp")
	for _, bounds := range [][2]float32{{-1, 1}, {0, 0}, {math.Float32frombits(0x80000000), 0}, {float32(math.Inf(-1)), 1}, {0, float32(math.Inf(1))}, {float32(math.Inf(1)), float32(math.Inf(1))}} {
		for _, pattern := range []string{"none", "all", "alternating", "clustered", "shuffled"} {
			o := DefaultOptions()
			o.Lower = math.Float32bits(bounds[0])
			o.Upper = math.Float32bits(bounds[1])
			o.Pattern = pattern
			o.Seed = 123
			if err := o.ValidateFor(op); err != nil {
				t.Fatal(err)
			}
			for _, n := range []int{0, 1, 7, 8, 9, 1025} {
				x, y, out := MakeForOperation(op, n, o)
				if len(x) != n || y != nil || len(out) != n || FingerprintFloat32(x, nil) != FingerprintForOperation(op, n, o) {
					t.Fatal("clamp arity/hash")
				}
				clipped := 0
				for _, v := range x {
					if v < bounds[0] || v > bounds[1] {
						clipped++
					}
				}
				switch pattern {
				case "none":
					if clipped != 0 {
						t.Fatal(clipped)
					}
				case "all":
					if clipped != n {
						t.Fatal(clipped, n)
					}
				case "alternating", "clustered":
					if clipped != (n+1)/2 {
						t.Fatal(clipped, n)
					}
				case "shuffled":
					if n == 1025 && (clipped < 400 || clipped > 625) {
						t.Fatal(clipped)
					}
				}
			}
		}
	}
	o := DefaultOptions()
	o.Lower = 0xff800000
	o.Upper = 0x7f800000
	if o.ValidateFor(op) == nil {
		t.Fatal("impossible clipped pattern accepted")
	}
	o.Pattern = "none"
	if e := o.ValidateFor(op); e != nil {
		t.Fatal(e)
	}
	for _, bits := range []uint32{0x7fc12345, 0x7f812345} {
		o.Lower = bits
		if o.ValidateFor(op) == nil {
			t.Fatal("NaN bound accepted")
		}
	}
	o = DefaultOptions()
	o.Pattern = "shuffled"
	o.Seed = 123
	before := o.Identity(op) + o.Generator(op) + FingerprintForOperation(op, 1025, o)
	o.Alpha++
	o.Beta++
	o.Threshold++
	o.SumInput = "cancellation"
	o.DotInput = "fusion-sensitive"
	o.FMAInput = "rounding"
	if before != o.Identity(op)+o.Generator(op)+FingerprintForOperation(op, 1025, o) {
		t.Fatal("irrelevant clamp controls")
	}
}
