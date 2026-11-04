package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
)

func main() {
	bits := []uint32{0, 0x80000000, 1, 0x7f800000, 0x7fc12345, 0x7f812345}
	x := make([]float32, len(bits))
	for i, b := range bits {
		x[i] = math.Float32frombits(b)
	}
	for _, kernel := range []struct {
		name string
		run  func([]float32, []float32) error
	}{{"Scalar", patterns.BitRoundTripScalar}, {"SIMDPartial", patterns.BitRoundTripSIMD}} {
		out := make([]float32, len(x))
		if e := kernel.run(x, out); e != nil {
			panic(e)
		}
		fmt.Printf("%s:", kernel.name)
		for _, v := range out {
			fmt.Printf(" %08x", math.Float32bits(v))
		}
		fmt.Println()
	}
	fmt.Println("Bits are reinterpreted, not numerically converted; timings do not prove bitcast instructions are emitted.")
}
