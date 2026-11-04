package main

import (
	"fmt"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

func main() {
	x := []float32{-1e20, 1, 1e20, 0, 1e20, 0, -1e20, 0, 2}
	out := make([]float32, (len(x)+7)/8)
	if e := reduction.Sum8SIMD(x, out); e != nil {
		panic(e)
	}
	fmt.Println(out) // [1 2]; last block padded with +0.
}
