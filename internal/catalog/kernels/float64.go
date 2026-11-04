package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

var float64Bindings = map[string]func(a, b, out []float64) func() error{
	catalog.Scalar:      func(a, b, out []float64) func() error { return func() error { return addition.Add64Scalar(a, b, out) } },
	catalog.SIMDPartial: func(a, b, out []float64) func() error { return func() error { return addition.Add64SIMD(a, b, out) } },
}
