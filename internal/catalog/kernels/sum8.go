package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

var sum8Bindings = map[string]func(x, out []float32) func() error{
	catalog.Scalar:      func(x, out []float32) func() error { return func() error { return reduction.Sum8Scalar(x, out) } },
	catalog.SIMDPartial: func(x, out []float32) func() error { return func() error { return reduction.Sum8SIMD(x, out) } },
}
