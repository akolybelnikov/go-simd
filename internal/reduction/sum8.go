package reduction

import "simd"

func sum8Outputs(n int) int {
	count := n / 8
	if n%8 != 0 {
		count++
	}
	return count
}

// Sum8Scalar returns one specified-order result per eight inputs, padding the
// last block with +0. Same-start output-prefix alias is supported; other overlap
// is outside the contract. Output extent errors precede writes.
func Sum8Scalar(x, out []float32) error {
	if len(out) != sum8Outputs(len(x)) {
		return ErrLengthMismatch
	}
	for start := 0; start < len(x); start += 8 {
		var block [8]float32
		copy(block[:], x[start:])
		p0 := float32(block[0] + block[4])
		p1 := float32(block[1] + block[5])
		p2 := float32(block[2] + block[6])
		p3 := float32(block[3] + block[7])
		out[start/8] = float32(float32(float32(p0+p1)+p2) + p3)
	}
	return nil
}

// Sum8SIMD composes portable lane-wise pair sums and an explicitly timed scalar
// final fold. The algorithm's eight-input block is independent of native width.
func Sum8SIMD(x, out []float32) error {
	if len(out) != sum8Outputs(len(x)) {
		return ErrLengthMismatch
	}
	var vector simd.Float32s
	w := vector.Len()
	for start := 0; start < len(x); start += 8 {
		var block [8]float32
		copy(block[:], x[start:])
		var pairs [4]float32
		for i := 0; i < 4; {
			end := i + min(w, 4-i)
			a, n := simd.LoadFloat32sPart(block[i:end])
			b, _ := simd.LoadFloat32sPart(block[i+4 : end+4])
			a.Add(b).StorePart(pairs[i:end])
			i += n
		}
		out[start/8] = float32(float32(float32(pairs[0]+pairs[1])+pairs[2]) + pairs[3])
	}
	return nil
}
