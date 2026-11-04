package input

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
)

func (o Options) muladdTriple(i int) (float32, float32, float32) {
	switch o.FMAFixture() {
	case "rounding":
		u := float32(0x1p-23)
		triples := [8][3]float32{{1 + u, 1 - u, -1}, {1 + u, 1 - u, 16777218}, {math.Float32frombits(0x80000000), 1, math.Float32frombits(0x80000000)}, {math.SmallestNonzeroFloat32, 0.5, math.SmallestNonzeroFloat32}, {math.MaxFloat32, 2, -math.MaxFloat32}, {1 + u, 1 - u, -1}, {-1 - u, 1 - u, 1}, {1, 1, -1}}
		p := triples[i%len(triples)]
		return p[0], p[1], p[2]
	case "mixed-magnitude":
		p := [8][3]float32{{1e20, 1e-20, -1}, {1e-20, 1e20, 1}, {math.MaxFloat32, 0.5, -1e20}, {1, math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat32}, {-1e20, 1e-20, 1}, {3, -3, 1}, {1e-20, 1e-20, -1e-20}, {-1, 1, 1}}[i%8]
		return p[0], p[1], p[2]
	default:
		return float32(i%33-16) / 4, float32(i%17-8) / 8, float32(i%19-9) / 16
	}
}

// MakeMulAdd allocates all three real operands and output outside timing.
func MakeMulAdd(n int, o Options) (a, b, c, out []float32) {
	a, b, c, out = make([]float32, n), make([]float32, n), make([]float32, n), make([]float32, n)
	for i := range a {
		a[i], b[i], c[i] = o.muladdTriple(i)
	}
	return
}

// FingerprintTernaryFloat32 hashes ordered a/b/c float32 bits, little endian.
// Operation arity and common element count are separately retained in identity.
func FingerprintTernaryFloat32(a, b, c []float32) string {
	h := sha256.New()
	for _, x := range [][]float32{a, b, c} {
		hashValues(h, len(x), func(i int) float32 { return x[i] })
	}
	return hex.EncodeToString(h.Sum(nil))
}
func fingerprintMulAdd(n int, o Options) string {
	h := sha256.New()
	for operand := 0; operand < 3; operand++ {
		hashValues(h, n, func(i int) float32 {
			a, b, c := o.muladdTriple(i)
			switch operand {
			case 0:
				return a
			case 1:
				return b
			default:
				return c
			}
		})
	}
	return hex.EncodeToString(h.Sum(nil))
}
