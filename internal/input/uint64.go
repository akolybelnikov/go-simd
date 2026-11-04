package input

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
)

const Clmul64GeneratorID = "generated-clmul-uint64-v1"

func clmul64Pair(i int) (uint64, uint64) {
	values := [...]uint64{0, 1, 3, 0x8000000000000000, 0xffffffffffffffff, 0x0123456789abcdef, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555}
	if i < len(values)*len(values) {
		return values[(i/len(values))%len(values)], values[i%len(values)]
	}
	x := uint64(i) + 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31), x ^ 0xd6e8feb86659fd93
}
func MakeClmul64(n int) (a, b, out []uint64) {
	a, b, out = make([]uint64, n), make([]uint64, n), make([]uint64, n)
	for i := range a {
		a[i], b[i] = clmul64Pair(i)
	}
	return
}
func hashUint64Values(h hash.Hash, n int, value func(int) uint64) {
	var block [4096]byte
	for start := 0; start < n; {
		count := min(n-start, len(block)/8)
		for j := range count {
			binary.LittleEndian.PutUint64(block[j*8:], value(start+j))
		}
		h.Write(block[:count*8])
		start += count
	}
}
func FingerprintUint64(a, b []uint64) string {
	h := sha256.New()
	for _, x := range [][]uint64{a, b} {
		hashUint64Values(h, len(x), func(i int) uint64 { return x[i] })
	}
	return hex.EncodeToString(h.Sum(nil))
}
func Clmul64Fingerprint(n int) string {
	h := sha256.New()
	hashUint64Values(h, n, func(i int) uint64 { a, _ := clmul64Pair(i); return a })
	hashUint64Values(h, n, func(i int) uint64 { _, b := clmul64Pair(i); return b })
	return hex.EncodeToString(h.Sum(nil))
}
