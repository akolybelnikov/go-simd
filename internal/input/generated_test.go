package input

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
)

func TestFingerprintEncoding(t *testing.T) {
	// Independent IEEE-754 little-endian encoding: 1, negative zero, NaN with
	// payload 0x123, then -2.5 in the second operand (not interleaved).
	a := []float32{1, math.Float32frombits(0x80000000), math.Float32frombits(0x7fc00123)}
	b := []float32{-2.5}
	bits := []byte{0, 0, 128, 63, 0, 0, 0, 128, 35, 1, 192, 127, 0, 0, 32, 192}
	want := sha256.Sum256(bits)
	if got := FingerprintFloat32(a, b); got != hex.EncodeToString(want[:]) {
		t.Fatalf("fingerprint = %s, want %x", got, want)
	}
	if FingerprintFloat32(a, b) == FingerprintFloat32(b, a) {
		t.Fatal("operand order must affect identity")
	}
}

func TestGeneratedFingerprintMatchesPreparedInputs(t *testing.T) {
	for _, size := range []int{0, 1, 49, 50, 99, 100, 101, 1024, 1025} {
		a, b, _ := MakeFloat32Slices(size)
		if got, want := GeneratedFloat32Fingerprint(size), FingerprintFloat32(a, b); got != want {
			t.Errorf("size %d: generated hash %s != prepared input hash %s", size, got, want)
		}
	}
	if got := GeneratedFloat32Fingerprint(0); got != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("empty hash = %s", got)
	}
}
