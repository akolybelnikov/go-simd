package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/movement"
)

var copy64Bindings = map[string]func(x, out []float64) func() error{
	catalog.Scalar:      func(x, out []float64) func() error { return func() error { return movement.Copy64Scalar(x, out) } },
	catalog.GoCopy:      func(x, out []float64) func() error { return func() error { return movement.Copy64Go(x, out) } },
	catalog.SIMDPartial: func(x, out []float64) func() error { return func() error { return movement.Copy64SIMD(x, out) } },
}
