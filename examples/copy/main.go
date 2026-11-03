// Run inside this module with GOEXPERIMENT=simd.
package main

import (
	"fmt"
	"math"

	"github.com/akolybelnikov/go-simd/internal/movement"
)

func main() {
	x := []float32{math.Float32frombits(0x80000000), math.Float32frombits(0x7f812345), math.Float32frombits(0x7fc12345), math.Float32frombits(1), 1}
	for _, k := range []struct {
		name string
		run  func([]float32, []float32) error
	}{{"Scalar", movement.CopyScalar}, {"GoCopy", movement.CopyBuiltin}, {"SIMDPartial", movement.CopySIMD}, {"SIMDTail", movement.CopySIMDVectorTail}} {
		out := make([]float32, len(x))
		if err := k.run(x, out); err != nil {
			panic(err)
		}
		bits := make([]uint32, len(out))
		for i, v := range out {
			bits[i] = math.Float32bits(v)
		}
		fmt.Printf("%s: %08x\n", k.name, bits)
	}
	if err := movement.CopySIMDVectorTail(x, x); err != nil {
		panic(err)
	}
	fmt.Printf("exact alias preserved first bits: %08x\n", math.Float32bits(x[0]))
}
