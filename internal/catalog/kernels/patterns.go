package kernels

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

// Scalar extraction/binding is prepared once, while each kernel's broadcasts
// and length validation occur per measured call. No vectors cross this seam.
var configuredBindings = map[bindingKey]func([]float32, []float32, []float32, input.Options) func() error{
	{catalog.Scale, catalog.Scalar}: func(a, b, out []float32, o input.Options) func() error {
		alpha := o.Float("alpha")
		return func() error { return patterns.ScaleScalar(a, out, alpha) }
	},
	{catalog.Scale, catalog.SIMDPartial}: func(a, b, out []float32, o input.Options) func() error {
		alpha := o.Float("alpha")
		return func() error { return patterns.ScaleSIMD(a, out, alpha) }
	},
	{catalog.Affine, catalog.Scalar}: func(a, b, out []float32, o input.Options) func() error {
		alpha, beta := o.Float("alpha"), o.Float("beta")
		return func() error { return patterns.AffineScalar(a, out, alpha, beta) }
	},
	{catalog.Affine, catalog.SIMDPartial}: func(a, b, out []float32, o input.Options) func() error {
		alpha, beta := o.Float("alpha"), o.Float("beta")
		return func() error { return patterns.AffineSIMD(a, out, alpha, beta) }
	},
	{catalog.ThresholdZero, catalog.Scalar}: func(a, b, out []float32, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.ThresholdZeroScalar(a, out, t) }
	},
	{catalog.ThresholdZero, catalog.SIMDPartial}: func(a, b, out []float32, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.ThresholdZeroSIMD(a, out, t) }
	},
	{catalog.SelectGT, catalog.Scalar}: func(a, b, out []float32, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.SelectGTScalar(a, b, out, t) }
	},
	{catalog.SelectGT, catalog.SIMDPartial}: func(a, b, out []float32, o input.Options) func() error {
		t := o.Float("threshold")
		return func() error { return patterns.SelectGTSIMD(a, b, out, t) }
	},
	{catalog.RangeZero, catalog.Scalar}: func(a, b, out []float32, o input.Options) func() error {
		lo, hi := o.Float("lower"), o.Float("upper")
		return func() error { return patterns.RangeZeroScalar(a, out, lo, hi) }
	},
	{catalog.RangeZero, catalog.SIMDPartial}: func(a, b, out []float32, o input.Options) func() error {
		lo, hi := o.Float("lower"), o.Float("upper")
		return func() error { return patterns.RangeZeroSIMD(a, out, lo, hi) }
	},
}
