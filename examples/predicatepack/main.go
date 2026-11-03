// Run inside this module with GOEXPERIMENT=simd.
package main

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/patterns"
)

func main() {
	x := []float32{1, 0, 0, 1, 0, 0, 1, 0, 1}
	plain := make([]uint8, len(x))
	scalar, vector := make([]uint8, (len(x)+7)/8), make([]uint8, (len(x)+7)/8)
	if err := patterns.PredicateGTScalar(x, plain, 0); err != nil {
		panic(err)
	}
	if err := patterns.PredicateGTPackedScalar(x, scalar, 0); err != nil {
		panic(err)
	}
	if err := patterns.PredicateGTPackedSIMD(x, vector, 0); err != nil {
		panic(err)
	}
	fmt.Printf("byte mask (%d bytes): %v\n", len(plain), plain)
	fmt.Printf("packed Scalar (%d bytes): %02x\n", len(scalar), scalar)
	fmt.Printf("packed SIMDPartial (%d bytes): %02x\n", len(vector), vector)
	if err := patterns.PredicateGTPackedSIMD([]float32{0}, vector[:1], -1); err != nil {
		panic(err)
	}
	fmt.Printf("one true input; unused bits zero: %08b\n", vector[0])
}
