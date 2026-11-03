package patterns

import "simd"

// PredicateGTPackedScalar exports ordered x > threshold into ceil(len(x)/8)
// bytes. Input i is bit i%8 of out[i/8], least-significant bit first. Unused
// high bits are zero. Length errors precede writes; storage must be disjoint.
func PredicateGTPackedScalar(x []float32, out []uint8, threshold float32) error {
	if len(out) != len(x)/8+(len(x)%8+7)/8 {
		return ErrLengthMismatch
	}
	for byteIndex := range out {
		var value uint8
		start := byteIndex * 8
		for bit := 0; bit < min(8, len(x)-start); bit++ {
			if x[start+bit] > threshold {
				value |= 1 << uint(bit)
			}
		}
		out[byteIndex] = value
	}
	return nil
}

// PredicateGTPackedSIMD compares/selects explicit 0/1 values and converts them
// into scratch before packing valid lanes. A partial output byte carries across
// vector boundaries. No SDK mask representation or padded input bit escapes.
func PredicateGTPackedSIMD(x []float32, out []uint8, threshold float32) error {
	if len(out) != len(x)/8+(len(x)%8+7)/8 {
		return ErrLengthMismatch
	}
	t := simd.BroadcastFloat32s(threshold)
	one, zero := simd.BroadcastFloat32s(1), simd.BroadcastFloat32s(0)
	var scratch [16]int32
	lanes := scratch[:]
	if w := t.Len(); w > len(lanes) {
		lanes = make([]int32, w)
	}
	var value uint8
	for i := 0; i < len(x); {
		v, n := simd.LoadFloat32sPart(x[i:])
		one.IfElse(v.Greater(t), zero).ConvertToInt32().Store(lanes)
		for j := 0; j < n; j++ {
			index := i + j
			value |= uint8(lanes[j]) << uint(index%8)
			if index%8 == 7 {
				out[index/8] = value
				value = 0
			}
		}
		i += n
	}
	if len(x)%8 != 0 {
		out[len(x)/8] = value
	}
	return nil
}
