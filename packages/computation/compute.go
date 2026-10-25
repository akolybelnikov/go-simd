package computation

import "simd"

// ComputeHeavyScalar evaluates a*a + b*b + a*b for each pair of values.
// Input and output slices must have equal lengths.
func ComputeHeavyScalar(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return BadInputError
	}
	for i := range a {
		x, y := a[i], b[i]
		out[i] = x*x + y*y + x*y
	}
	return nil
}

// ComputeHeavySIMD evaluates a*a + b*b + a*b using SIMD vectors.
// Input and output slices must have equal lengths.
func ComputeHeavySIMD(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return BadInputError
	}
	for i := 0; i < len(a); {
		x, n := simd.LoadFloat32sPart(a[i:])
		y, _ := simd.LoadFloat32sPart(b[i:])
		x.Mul(x).Add(y.Mul(y)).Add(x.Mul(y)).StorePart(out[i:])
		i += n
	}
	return nil
}
