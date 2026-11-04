package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

var ternaryBindings = map[string]func(a, b, c, out []float32) func() error{
	catalog.Scalar: func(a, b, c, out []float32) func() error {
		return func() error { return patterns.MulAddScalar(a, b, c, out) }
	},
	catalog.SIMDPartial: func(a, b, c, out []float32) func() error {
		return func() error { return patterns.MulAddSIMD(a, b, c, out) }
	},
}
