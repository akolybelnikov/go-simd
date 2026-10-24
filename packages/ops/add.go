package ops

import (
	"simd"
)

func Add(a, b, out []float32) {
	if len(a) != len(b) || len(a) != len(out) {
		a, b, out = PadSlices(a, b, out)
	}
	for i := 0; i < len(a); {
		x, n := simd.LoadFloat32sPart(a[i:])
		y, _ := simd.LoadFloat32sPart(b[i:])
		x.Add(y).StorePart(out[i:])
		i += n
	}
}

func padSlice(a []float32, maxLen int) []float32 {
	if len(a) < maxLen {
		padded := make([]float32, maxLen)
		copy(padded, a)
		return padded
	}
	return a
}

func PadSlices(a, b, out []float32) ([]float32, []float32, []float32) {
	maxLen := max(len(a), len(b), len(out))
	return padSlice(a, maxLen), padSlice(b, maxLen), padSlice(out, maxLen)
}
