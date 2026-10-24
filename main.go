package main

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/packages/ops"
)

func main() {
	a := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}
	b := []float32{5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0}
	out := make([]float32, len(a))
	ops.AddSIMD(a, b, out)
	fmt.Printf("%v", out)
}
