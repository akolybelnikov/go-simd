package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"simd"
)

func main() {
	a, b := []uint8{0, 128, 254, 255}, []uint8{1, 128, 2, 1}
	var vector simd.Uint8s
	fmt.Printf("Uint8 lanes=%d SDK emulated=%t\n", vector.Len(), simd.Emulated())
	for _, kernel := range []struct {
		name string
		run  func([]uint8, []uint8, []uint8) error
	}{{"Scalar", integer.SaturatingAdd8Scalar}, {"SIMDPartial", integer.SaturatingAdd8SIMD}} {
		out := make([]uint8, len(a))
		if e := kernel.run(a, b, out); e != nil {
			panic(e)
		}
		fmt.Println(kernel.name, out)
	}
	wrap := make([]uint8, len(a))
	for i := range a {
		wrap[i] = a[i] + b[i]
	}
	fmt.Println("Wrapping differs (not a registered control):", wrap)
}
