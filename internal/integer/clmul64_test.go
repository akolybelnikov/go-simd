package integer

import (
	"errors"
	"math/bits"
	"math/rand/v2"
	"simd"
	"slices"
	"testing"
)

// Coefficient convolution is independent of the kernel shift-and-XOR loop.
func polynomialLow64(a, b uint64) uint64 {
	var out uint64
	for power := 0; power < 64; power++ {
		odd := false
		for left := max(0, power-bits.Len64(b)+1); left <= power && left < bits.Len64(a); left++ {
			right := power - left
			if a&(uint64(1)<<left) != 0 && b&(uint64(1)<<right) != 0 {
				odd = !odd
			}
		}
		if odd {
			out |= uint64(1) << power
		}
	}
	return out
}
func TestClmul64PolynomialTruth(t *testing.T) {
	var v simd.Uint64s
	w := v.Len()
	rng := rand.New(rand.NewPCG(31, 57))
	for _, fn := range []func([]uint64, []uint64, []uint64) error{Clmul64Scalar, Clmul64SIMD} {
		a, b := make([]uint64, 65536), make([]uint64, 65536)
		for i := range a {
			a[i] = uint64(i / 256)
			b[i] = uint64(i % 256)
		}
		out := make([]uint64, len(a))
		if e := fn(a, b, out); e != nil {
			t.Fatal(e)
		}
		for i := range out {
			if out[i] != polynomialLow64(a[i], b[i]) {
				t.Fatal("8-bit exhaustive", i, out[i])
			}
		}
		for _, n := range []int{0, 1, 2, 3, w - 1, w, w + 1, 2*w + 1, 1025} {
			a, b := make([]uint64, n), make([]uint64, n)
			for i := range a {
				a[i], b[i] = rng.Uint64(), rng.Uint64()
			}
			if n > 0 {
				a[0] = uint64(1) << 63
				b[0] = 3
			}
			beforeA, beforeB := slices.Clone(a), slices.Clone(b)
			expected := make([]uint64, n)
			for i := range a {
				expected[i] = polynomialLow64(a[i], b[i])
			}
			storage := make([]uint64, n+2)
			storage[0] = 123
			storage[n+1] = 456
			out := storage[1 : n+1]
			if e := fn(a, b, out); e != nil || !slices.Equal(out, expected) || storage[0] != 123 || storage[n+1] != 456 || !slices.Equal(a, beforeA) || !slices.Equal(b, beforeB) {
				t.Fatal("boundary/bit truth", n, e, out)
			}
			if e := fn(a, b, a); e != nil || !slices.Equal(a, expected) {
				t.Fatal("alias a", n, e)
			}
			copy(a, beforeA)
			if e := fn(a, b, b); e != nil || !slices.Equal(b, expected) {
				t.Fatal("alias b", n, e)
			}
		}
		out = make([]uint64, 4)
		for i := range out {
			out[i] = 99
		}
		if e := fn([]uint64{1}, []uint64{1}, out); !errors.Is(e, ErrLengthMismatch) || out[0] != 99 {
			t.Fatal("error before write", e)
		}
		x, y, z := make([]uint64, 17), make([]uint64, 17), make([]uint64, 17)
		if got := testing.AllocsPerRun(10, func() {
			if e := fn(x, y, z); e != nil {
				panic(e)
			}
		}); got != 0 {
			t.Fatal("allocation", got)
		}
	}
	if polynomialLow64(3, 3) != 5 {
		t.Fatal("oracle")
	}
}
