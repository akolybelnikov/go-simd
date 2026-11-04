package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

var bitRoundTripBindings = map[string]func(x, out []float32) func() error{
	catalog.Scalar: func(x, out []float32) func() error {
		return func() error { return patterns.BitRoundTripScalar(x, out) }
	},
	catalog.SIMDPartial: func(x, out []float32) func() error { return func() error { return patterns.BitRoundTripSIMD(x, out) } },
}
