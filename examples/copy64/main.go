package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/movement"
	"math"
)

func main() {
	x := []float64{math.Float64frombits(1 << 63), math.Float64frombits(0x7ff0123456789abc)}
	out := make([]float64, len(x))
	if e := movement.Copy64SIMD(x, out); e != nil {
		panic(e)
	}
	for _, v := range out {
		fmt.Printf("%016x\n", math.Float64bits(v))
	}
}
