package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
)

func main() {
	a := []float32{1, -2, math.Float32frombits(0x7f812345)}
	masks := []float32{math.Float32frombits(0x80000000), math.Float32frombits(0x80000000), 0}
	for _, kernel := range []struct {
		name string
		run  func([]float32, []float32, []float32) error
	}{{"Scalar", patterns.XorBitsScalar}, {"SIMDPartial", patterns.XorBitsSIMD}} {
		out := make([]float32, len(a))
		if e := kernel.run(a, masks, out); e != nil {
			panic(e)
		}
		fmt.Printf("%s:", kernel.name)
		for _, v := range out {
			fmt.Printf(" %08x", math.Float32bits(v))
		}
		fmt.Println()
	}
	fmt.Println("A sign-bit mask flips 1 to -1 and -2 to 2; zero mask retains the NaN payload. No floating arithmetic occurs.")
}
