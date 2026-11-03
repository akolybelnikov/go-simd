package input

import "math"

const CopyGeneratorID = "generated-copy-float32-bits-v1"

func copyValue(i int) float32 {
	bits := [16]uint32{0, 0x80000000, 0x3f800000, 0xbf800000, 0x7f800000, 0xff800000, 0x7fc12345, 0x7f812345, 1, 0x80000001, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff, 0x3eaaaaab, 0xdeadbeef}
	return math.Float32frombits(bits[i%len(bits)])
}
