// Command basic demonstrates module-local kernel use without benchmarking.
package main

import (
	"errors"
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/computation"
)

func main() {
	a, b := []float32{1, 2, 3}, []float32{4, 5, 6}
	out := make([]float32, len(a))
	for _, kernel := range []struct {
		name string
		run  func([]float32, []float32, []float32) error
	}{
		{"addition/Scalar", addition.AddScalar},
		{"addition/SIMDPartial", addition.AddSIMD},
		{"addition/SIMDTail", addition.AddSIMDVectorTail},
		{"addition/SIMDParallel", addition.AddSIMDParallel},
		{"computation/Scalar", computation.ComputeHeavyScalar},
		{"computation/SIMD", computation.ComputeHeavySIMD},
	} {
		must(kernel.run(a, b, out))
		fmt.Printf("%s: %v\n", kernel.name, out)
	}

	aliasA := append([]float32(nil), a...)
	must(addition.AddSIMD(aliasA, b, aliasA))
	fmt.Printf("addition/output aliases a: %v\n", aliasA)
	aliasB := append([]float32(nil), b...)
	must(addition.AddSIMD(a, aliasB, aliasB))
	fmt.Printf("addition/output aliases b: %v\n", aliasB)

	unchanged := []float32{123}
	err := addition.AddSIMD(a, b, unchanged)
	fmt.Printf("addition/length mismatch: %t; output=%v\n", errors.Is(err, addition.ErrLengthMismatch), unchanged)
	must(addition.AddSIMD(nil, nil, nil))
	fmt.Println("addition/empty: ok")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
