package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

var predicateBindings = map[string]func([]float32, []uint8, input.Options) func() error{
	catalog.Scalar: func(x []float32, out []uint8, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.PredicateGTScalar(x, out, t) }
	},
	catalog.SIMDPartial: func(x []float32, out []uint8, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.PredicateGTSIMD(x, out, t) }
	},
}
