package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/integer"
)

var uint8Bindings = map[string]func(a, b, out []uint8) func() error{
	catalog.Scalar: func(a, b, out []uint8) func() error {
		return func() error { return integer.SaturatingAdd8Scalar(a, b, out) }
	},
	catalog.SIMDPartial: func(a, b, out []uint8) func() error {
		return func() error { return integer.SaturatingAdd8SIMD(a, b, out) }
	},
}
