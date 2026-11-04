// Package kernels binds catalogue metadata to the existing typed Go kernels.
// SIMD dependencies stay here rather than in the metadata/result packages.
package kernels

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/computation"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
)

type Float32Kernel func(a, b, out []float32) error
type Float32Binder func(a, b, out []float32) func() error

type Implementation struct {
	BindByteArrays    func(a, b, out []uint8) func() error
	BindFloat64       func(a, b, out []float64) func() error
	BindTernary       func(a, b, c, out []float32) func() error
	BindUnary         func(x, out []float32) func() error
	BindUint8         func(x []float32, out []uint8, options input.Options) func() error
	BindCheckedScalar func(x, y []float32) func() (float32, error)
	BindScalar        func(x []float32) func() float32
	BindConfigured    func(a, b, out []float32, options input.Options) func() error
	Descriptor        catalog.Variant
	Run               Float32Kernel
	Bind              Float32Binder
}

type Operation struct {
	Descriptor      catalog.Operation
	Implementations []Implementation
}

type bindingKey struct{ operation, strategy string }
type binding struct {
	run  Float32Kernel
	bind Float32Binder
}

// Each explicit binder preserves the quick CLI's direct kernel call inside one
// bound closure. Generic binding through a function variable would introduce an
// extra indirect dispatch in that timed path. Native benches use run directly.
var bindings = map[bindingKey]binding{
	{catalog.Add, catalog.Scalar}: {addition.AddScalar, func(a, b, out []float32) func() error {
		return func() error { return addition.AddScalar(a, b, out) }
	}},
	{catalog.Add, catalog.SIMDPartial}: {addition.AddSIMD, func(a, b, out []float32) func() error {
		return func() error { return addition.AddSIMD(a, b, out) }
	}},
	{catalog.Add, catalog.SIMDScalarTail}: {addition.AddSIMDVectorTail, func(a, b, out []float32) func() error {
		return func() error { return addition.AddSIMDVectorTail(a, b, out) }
	}},
	{catalog.Add, catalog.SIMDParallelScalarTail}: {addition.AddSIMDParallel, func(a, b, out []float32) func() error {
		return func() error { return addition.AddSIMDParallel(a, b, out) }
	}},
	{catalog.ComputeExpression, catalog.Scalar}: {computation.ComputeHeavyScalar, func(a, b, out []float32) func() error {
		return func() error { return computation.ComputeHeavyScalar(a, b, out) }
	}},
	{catalog.ComputeExpression, catalog.SIMDPartial}: {computation.ComputeHeavySIMD, func(a, b, out []float32) func() error {
		return func() error { return computation.ComputeHeavySIMD(a, b, out) }
	}},
}

// Lookup rejects unavailable operation/type combinations without casts/fallbacks.
// Resolution and buffer binding happen outside measured loops.
func Lookup(id, typeID string) (Operation, error) {
	if typeID != catalog.Float32 && typeID != catalog.Float64 && typeID != catalog.Uint8 {
		return Operation{}, fmt.Errorf("unavailable float32 binding for %q/%q", id, typeID)
	}
	descriptor, ok := catalog.Lookup(id, typeID)
	if !ok {
		return Operation{}, fmt.Errorf("unavailable operation/type %q/%q", id, typeID)
	}
	operation := Operation{Descriptor: descriptor}
	for _, variant := range descriptor.Variants {
		if descriptor.ID == catalog.SaturatingAdd8 {
			bind, ok := uint8Bindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing uint8 binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindByteArrays: bind})
			continue
		}
		if descriptor.ID == catalog.Add64 {
			bind, ok := float64Bindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing float64 binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindFloat64: bind})
			continue
		}
		if descriptor.ID == catalog.MulAdd {
			bind, ok := ternaryBindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing ternary binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindTernary: bind})
			continue
		}
		if descriptor.ID == catalog.Copy {
			bind, ok := copyBindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing copy binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindUnary: bind})
			continue
		}
		if descriptor.ID == catalog.PredicateGTPacked {
			bind, ok := packedPredicateBindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing packed predicate binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindUint8: bind})
			continue
		}
		if descriptor.ID == catalog.PredicateGT {
			bind, ok := predicateBindings[variant.ID]
			if !ok {
				return Operation{}, fmt.Errorf("missing predicate binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindUint8: bind})
			continue
		}
		if descriptor.ID == catalog.DotFMA {
			var bind func([]float32, []float32) func() (float32, error)
			switch variant.ID {
			case catalog.Scalar:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotFMAScalar(x, y) }
				}
			case catalog.SIMDStriped:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotFMASIMD(x, y) }
				}
			case catalog.SIMDStriped2:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotFMASIMD2(x, y) }
				}
			case catalog.SIMDStriped4:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotFMASIMD4(x, y) }
				}
			default:
				return Operation{}, fmt.Errorf("missing fused-dot binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindCheckedScalar: bind})
			continue
		}
		if descriptor.ID == catalog.Dot {
			var bind func([]float32, []float32) func() (float32, error)
			switch variant.ID {
			case catalog.Scalar:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotScalar(x, y) }
				}
			case catalog.SIMDStriped:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotSIMD(x, y) }
				}
			case catalog.SIMDStriped2:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotSIMD2(x, y) }
				}
			case catalog.SIMDStriped4:
				bind = func(x, y []float32) func() (float32, error) {
					return func() (float32, error) { return reduction.DotSIMD4(x, y) }
				}
			default:
				return Operation{}, fmt.Errorf("missing dot binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindCheckedScalar: bind})
			continue
		}
		if descriptor.ID == catalog.Sum {
			var bind func([]float32) func() float32
			switch variant.ID {
			case catalog.Scalar:
				bind = func(x []float32) func() float32 { return func() float32 { return reduction.SumScalar(x) } }
			case catalog.SIMDStriped:
				bind = func(x []float32) func() float32 { return func() float32 { return reduction.SumSIMD(x) } }
			case catalog.SIMDStriped2:
				bind = func(x []float32) func() float32 { return func() float32 { return reduction.SumSIMD2(x) } }
			case catalog.SIMDStriped4:
				bind = func(x []float32) func() float32 { return func() float32 { return reduction.SumSIMD4(x) } }
			default:
				return Operation{}, fmt.Errorf("missing sum binding for %s", variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindScalar: bind})
			continue
		}
		if descriptor.Parameterized() {
			bind, ok := configuredBindings[bindingKey{id, variant.ID}]
			if !ok {
				return Operation{}, fmt.Errorf("missing configured binding for %s/%s", id, variant.ID)
			}
			operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, BindConfigured: bind})
			continue
		}
		bound, ok := bindings[bindingKey{id, variant.ID}]
		if !ok || bound.run == nil || bound.bind == nil {
			return Operation{}, fmt.Errorf("missing kernel binding for %s/%s", id, variant.ID)
		}
		operation.Implementations = append(operation.Implementations, Implementation{Descriptor: variant, Run: bound.run, Bind: bound.bind})
	}
	return operation, nil
}
