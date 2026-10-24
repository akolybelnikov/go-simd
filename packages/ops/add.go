package ops

import (
	"simd"
)

func Add() {
	var v simd.Float32s
	n := v.Len()
	println("n", n)
}
