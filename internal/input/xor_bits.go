package input

import "math"

const XorBitsGeneratorID = "generated-xor-float32-bits-v1"

func xorBitsPair(i int) (float32, float32) {
	masks := [8]uint32{0, 0x80000000, 0x7fffffff, 0xffffffff, 0x007fffff, 0x7f800000, 0x55555555, 0xaaaaaaaa}
	return copyValue(i), math.Float32frombits(masks[i%len(masks)])
}
func MakeXorBits(n int) (a, b, out []float32) {
	a, b, out = make([]float32, n), make([]float32, n), make([]float32, n)
	for i := range a {
		a[i], b[i] = xorBitsPair(i)
	}
	return
}
