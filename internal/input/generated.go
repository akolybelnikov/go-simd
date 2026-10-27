// Package input prepares deterministic laboratory inputs outside measurement.
package input

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"math"
)

const GeneratedFloat32ID = "generated-float32-v1"

func generatedPair(i int) (float32, float32) {
	return float32(i%100) / 10, float32(i%50-25) / 10
}

// MakeFloat32Slices preserves the legacy generated inputs and reusable output.
func MakeFloat32Slices(size int) (a, b, out []float32) {
	a, b, out = make([]float32, size), make([]float32, size), make([]float32, size)
	for i := range a {
		a[i], b[i] = generatedPair(i)
	}
	return a, b, out
}

// FingerprintFloat32 hashes a's bits followed by b's bits in little-endian order.
// Signed zeros and NaN payloads are retained. Call only outside timing.
func FingerprintFloat32(a, b []float32) string {
	h := sha256.New()
	hashValues(h, len(a), func(i int) float32 { return a[i] })
	hashValues(h, len(b), func(i int) float32 { return b[i] })
	return hex.EncodeToString(h.Sum(nil))
}

// GeneratedFloat32Fingerprint computes the same hash without allocating slices.
func GeneratedFloat32Fingerprint(size int) string {
	return fingerprintGenerated(size, true, generatedPair)
}

func fingerprintGenerated(size int, binary bool, pair func(int) (float32, float32)) string {
	h := sha256.New()
	hashValues(h, size, func(i int) float32 { a, _ := pair(i); return a })
	if binary {
		hashValues(h, size, func(i int) float32 { _, b := pair(i); return b })
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashValues(h hash.Hash, size int, value func(int) float32) {
	var block [4096]byte
	for start := 0; start < size; {
		n := min(size-start, len(block)/4)
		for j := range n {
			binary.LittleEndian.PutUint32(block[j*4:], math.Float32bits(value(start+j)))
		}
		h.Write(block[:n*4]) // hash.Hash.Write never returns an error.
		start += n
	}
}
