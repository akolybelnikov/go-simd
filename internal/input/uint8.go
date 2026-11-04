package input

import (
	"crypto/sha256"
	"encoding/hex"
)

const SaturatingAdd8GeneratorID = "generated-saturating-add-uint8-v1"

func saturatingAdd8Pair(i int) (uint8, uint8) {
	p := [16][2]uint8{{0, 0}, {255, 0}, {255, 1}, {128, 127}, {128, 128}, {254, 1}, {254, 2}, {255, 255}, {1, 1}, {0, 255}, {127, 127}, {127, 129}, {200, 100}, {250, 10}, {64, 64}, {1, 254}}[i%16]
	return p[0], p[1]
}
func MakeSaturatingAdd8(n int) (a, b, out []uint8) {
	a, b, out = make([]uint8, n), make([]uint8, n), make([]uint8, n)
	for i := range a {
		a[i], b[i] = saturatingAdd8Pair(i)
	}
	return
}

// FingerprintUint8 hashes complete ordered byte operands outside timing.
func FingerprintUint8(a, b []uint8) string {
	h := sha256.New()
	h.Write(a)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
func SaturatingAdd8Fingerprint(n int) string {
	h := sha256.New()
	var block [4096]byte
	for operand := 0; operand < 2; operand++ {
		for start := 0; start < n; {
			count := min(n-start, len(block))
			for j := range count {
				a, b := saturatingAdd8Pair(start + j)
				block[j] = a
				if operand == 1 {
					block[j] = b
				}
			}
			h.Write(block[:count])
			start += count
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}
