package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/movement"
)

var copyBindings = map[string]func([]float32, []float32) func() error{
	catalog.Scalar:      func(x, out []float32) func() error { return func() error { return movement.CopyScalar(x, out) } },
	catalog.GoCopy:      func(x, out []float32) func() error { return func() error { return movement.CopyBuiltin(x, out) } },
	catalog.SIMDPartial: func(x, out []float32) func() error { return func() error { return movement.CopySIMD(x, out) } },
	catalog.SIMDScalarTail: func(x, out []float32) func() error {
		return func() error { return movement.CopySIMDVectorTail(x, out) }
	},
}
