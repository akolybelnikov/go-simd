package input

import (
	"math"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestCopyDeclaredBuffersBitsAndIdentity(t *testing.T) {
	op, _ := catalog.ByGroup("copy")
	o := DefaultOptions()
	want := []uint32{0, 0x80000000, 0x3f800000, 0xbf800000, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 1, 0x80000001, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff, 0x3eaaaaab, 0xdeadbeef}
	for _, n := range []int{0, 1, 17, 1025} {
		x, y, out := MakeForOperation(op, n, o)
		if y != nil || len(x) != n || len(out) != n || FingerprintFloat32(x, nil) != FingerprintForOperation(op, n, o) || o.Generator(op) != CopyGeneratorID || o.Identity(op) != "" {
			t.Fatal("copy arity/fingerprint")
		}
		for i, v := range x {
			if math.Float32bits(v) != want[i%16] {
				t.Fatal("copy fixture bits", i)
			}
		}
		if n > 0 {
			out[0] = 17
			if math.Float32bits(x[0]) != 0 {
				t.Fatal("copy fixture aliases output")
			}
		}
	}
	before := FingerprintForOperation(op, 1025, o)
	o.Alpha++
	o.Beta++
	o.Threshold = 0x7fc12345
	o.Lower++
	o.Upper++
	o.Pattern = "all"
	o.Seed = 123
	o.SumInput = "cancellation"
	o.DotInput = "fusion-sensitive"
	o.FMAInput = "rounding"
	if err := o.ValidateFor(op); err != nil {
		t.Fatal(err)
	}
	if FingerprintForOperation(op, 1025, o) != before || o.Identity(op) != "" || o.Generator(op) != CopyGeneratorID {
		t.Fatal("irrelevant options changed copy")
	}
}
