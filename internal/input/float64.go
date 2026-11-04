package input

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math"
)

const Addition64GeneratorID = "generated-addition-float64-v1"

func addition64Pair(i int) (float64, float64) { return float64(i%100) / 10, float64(i%50-25) / 10 }
func MakeAddition64(n int) (a, b, out []float64) {
	a, b, out = make([]float64, n), make([]float64, n), make([]float64, n)
	for i := range a {
		a[i], b[i] = addition64Pair(i)
	}
	return
}

// FingerprintFloat64 retains complete little-endian binary64 bits outside timing.
func FingerprintFloat64(a, b []float64) string {
	h := sha256.New()
	for _, x := range [][]float64{a, b} {
		hashFloat64Values(h, len(x), func(i int) float64 { return x[i] })
	}
	return hex.EncodeToString(h.Sum(nil))
}
func Addition64Fingerprint(n int) string {
	h := sha256.New()
	hashFloat64Values(h, n, func(i int) float64 { a, _ := addition64Pair(i); return a })
	hashFloat64Values(h, n, func(i int) float64 { _, b := addition64Pair(i); return b })
	return hex.EncodeToString(h.Sum(nil))
}
func hashFloat64Values(h hash.Hash, n int, value func(int) float64) {
	var block [4096]byte
	for start := 0; start < n; {
		count := min(n-start, len(block)/8)
		for j := range count {
			binary.LittleEndian.PutUint64(block[j*8:], math.Float64bits(value(start+j)))
		}
		h.Write(block[:count*8])
		start += count
	}
}
