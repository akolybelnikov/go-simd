package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/addition"
	"math"
	"simd"
)

func main() {
	var f32 simd.Float32s
	var f64 simd.Float64s
	fmt.Printf("lanes float32=%d float64=%d; SDK emulated=%t\n", f32.Len(), f64.Len(), simd.Emulated())
	a := []float64{16777216, 1, math.SmallestNonzeroFloat64}
	b := []float64{1, 0x1p-53, math.SmallestNonzeroFloat64}
	for _, kernel := range []struct {
		name string
		run  func([]float64, []float64, []float64) error
	}{{"Scalar", addition.Add64Scalar}, {"SIMDPartial", addition.Add64SIMD}} {
		out := make([]float64, len(a))
		if e := kernel.run(a, b, out); e != nil {
			panic(e)
		}
		fmt.Printf("%s: %v\n", kernel.name, out)
	}
	fmt.Println("binary32 at the first pair:", float32(a[0])+float32(b[0]), "; binary64:", a[0]+b[0])
	fmt.Println("At equal element counts, float64 addition accounts for twice the logical bytes of float32.")
}
