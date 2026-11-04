package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/integer"
)

var clmul64Bindings = map[string]func(a, b, out []uint64) func() error{
	catalog.Scalar:      func(a, b, out []uint64) func() error { return func() error { return integer.Clmul64Scalar(a, b, out) } },
	catalog.SIMDPartial: func(a, b, out []uint64) func() error { return func() error { return integer.Clmul64SIMD(a, b, out) } },
}
