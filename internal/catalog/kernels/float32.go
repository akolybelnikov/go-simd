// Package kernels binds catalogue metadata to the existing typed Go kernels.
// SIMD dependencies stay here rather than in the metadata/result packages.
package kernels

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/computation"
	"github.com/akolybelnikov/go-simd/internal/input"
)

type Float32Kernel func(a, b, out []float32) error
type Float32Binder func(a, b, out []float32) func() error

type Implementation struct {
	BindConfigured func(a, b, out []float32, options input.Options) func() error
	Descriptor     catalog.Variant
	Run            Float32Kernel
	Bind           Float32Binder
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
	if typeID != catalog.Float32 {
		return Operation{}, fmt.Errorf("unavailable float32 binding for %q/%q", id, typeID)
	}
	descriptor, ok := catalog.Lookup(id, typeID)
	if !ok {
		return Operation{}, fmt.Errorf("unavailable operation/type %q/%q", id, typeID)
	}
	operation := Operation{Descriptor: descriptor}
	for _, variant := range descriptor.Variants {
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
