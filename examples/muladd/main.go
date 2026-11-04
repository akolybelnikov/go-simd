package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"math"
)

func main() {
	a := []float32{1 + 0x1p-23, math.MaxFloat32, 2}
	b := []float32{1 - 0x1p-23, 2, 3}
	c := []float32{-1, -math.MaxFloat32, 4}
	fmt.Println("actual SIMD path:", patterns.FMAPath())
	for _, kernel := range []struct {
		name string
		run  func([]float32, []float32, []float32, []float32) error
	}{{"Scalar (software FMA32)", patterns.MulAddScalar}, {"SIMDPartial (verified gate)", patterns.MulAddSIMD}} {
		out := make([]float32, len(a))
		if e := kernel.run(a, b, c, out); e != nil {
			panic(e)
		}
		fmt.Printf("%s: %v; bits:", kernel.name, out)
		for _, v := range out {
			fmt.Printf(" %08x", math.Float32bits(v))
		}
		fmt.Println()
	}
	fmt.Println("The first fused result differs from a separately rounded multiply/add; the second avoids intermediate overflow.")
}
