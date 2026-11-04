package input

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
)

const Copy64GeneratorID = "generated-copy-float64-bits-v1"

func copy64Value(i int) float64 {
	words := [...]uint64{0, 0x8000000000000000, 1, 0x8000000000000001, 0x7ff0000000000000, 0xfff0000000000000, 0x7ff8123456789abc, 0xfff8123456789abc, 0x7ff0123456789abc, 0xfff0123456789abc, 0x7fefffffffffffff, 0x0010000000000000, 0x000fffffffffffff}
	if i < len(words) {
		return math.Float64frombits(words[i])
	}
	x := uint64(i)*0x9e3779b97f4a7c15 + 0xd6e8feb86659fd93
	return math.Float64frombits(x ^ (x >> 29))
}
func MakeCopy64(n int) (x, out []float64) {
	x, out = make([]float64, n), make([]float64, n)
	for i := range x {
		x[i] = copy64Value(i)
	}
	return
}
func Copy64Fingerprint(n int) string {
	h := sha256.New()
	hashFloat64Values(h, n, copy64Value)
	return hex.EncodeToString(h.Sum(nil))
}
