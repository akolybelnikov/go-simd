package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

var xorBitsBindings = map[string]binding{
	catalog.Scalar: {run: patterns.XorBitsScalar, bind: func(a, b, out []float32) func() error {
		return func() error { return patterns.XorBitsScalar(a, b, out) }
	}},
	catalog.SIMDPartial: {run: patterns.XorBitsSIMD, bind: func(a, b, out []float32) func() error { return func() error { return patterns.XorBitsSIMD(a, b, out) } }},
}
