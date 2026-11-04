package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"simd"
)

func main() {
	out := make([]uint64, 3)
	if e := integer.Clmul64SIMD([]uint64{3, 1 << 63, 1}, []uint64{3, 3, 7}, out); e != nil {
		panic(e)
	}
	fmt.Printf("low64 polynomial products: %x; SDK hardware capability: %t\n", out, simd.HasHardwareCarrylessMultiply())
}
