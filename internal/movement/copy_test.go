package movement_test

import (
	"errors"
	"math"
	"math/rand/v2"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/movement"
)

var copyKernels = []struct {
	name string
	run  func([]float32, []float32) error
}{
	{"Scalar", movement.CopyScalar}, {"GoCopy", movement.CopyBuiltin}, {"SIMDPartial", movement.CopySIMD}, {"SIMDTail", movement.CopySIMDVectorTail},
}

func TestCopyExactBitsAndBoundaries(t *testing.T) {
	var v simd.Float32s
	w := v.Len()
	rng := rand.New(rand.NewPCG(704, 1103))
	special := []uint32{0, 0x80000000, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 0xff812345, 1, 0x80000001, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff}
	for _, n := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025, 65537} {
		for _, offset := range []int{0, 1, 3} {
			xStorage := make([]float32, n+offset+1)
			x := xStorage[offset : offset+n : offset+n]
			want := make([]uint32, n)
			for i := range x {
				bits := rng.Uint32()
				if i < len(special) {
					bits = special[i]
				}
				want[i] = bits
				x[i] = math.Float32frombits(bits)
			}
			for _, k := range copyKernels {
				for _, alias := range []bool{false, true} {
					guard := make([]float32, n+offset+2)
					for i := range guard {
						guard[i] = math.Float32frombits(0xdeadbeef)
					}
					out := guard[offset+1 : offset+1+n : offset+1+n]
					if alias {
						out = x
					}
					if err := k.run(x, out); err != nil {
						t.Fatal(k.name, n, err)
					}
					for i, v := range out {
						if math.Float32bits(v) != want[i] || math.Float32bits(x[i]) != want[i] {
							t.Fatalf("%s n=%d alias=%v offset=%d index=%d: bits changed", k.name, n, alias, offset, i)
						}
					}
					if !alias && (math.Float32bits(guard[offset]) != 0xdeadbeef || math.Float32bits(guard[offset+1+n]) != 0xdeadbeef) {
						t.Fatal("output canary overwritten", k.name)
					}
				}
			}
		}
	}
}

func TestCopyLengthEmptyAndAllocations(t *testing.T) {
	for _, k := range copyKernels {
		if err := k.run(nil, nil); err != nil {
			t.Fatal(k.name, err)
		}
		for _, pair := range [][2]int{{0, 1}, {1, 0}, {3, 2}, {2, 3}} {
			out := make([]float32, pair[1])
			for i := range out {
				out[i] = math.Float32frombits(0x7f812345)
			}
			if err := k.run(make([]float32, pair[0]), out); !errors.Is(err, movement.ErrLengthMismatch) {
				t.Fatal(k.name, err)
			}
			for _, v := range out {
				if math.Float32bits(v) != 0x7f812345 {
					t.Fatal("length failure wrote output", k.name)
				}
			}
		}
		x, out := make([]float32, 1025), make([]float32, 1025)
		if a := testing.AllocsPerRun(20, func() {
			if err := k.run(x, out); err != nil {
				panic(err)
			}
		}); a != 0 {
			t.Fatal("copy allocation", k.name, a)
		}
	}
}
