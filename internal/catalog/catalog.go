// Package catalog defines the laboratory's implemented operations and identities.
// Metadata is independent of SIMD, testing, execution, and report rendering.
package catalog

import (
	"fmt"
	"math"
	"slices"
)

const (
	Float32                = "float32"
	Add                    = "add"
	ComputeExpression      = "compute-expression"
	Scalar                 = "scalar"
	SIMDPartial            = "simd-partial"
	SIMDScalarTail         = "simd-scalar-tail"
	SIMDParallelScalarTail = "simd-parallel-scalar-tail"
	KernelOnly             = "kernel-only"
)

type Shape string

const (
	Array       Shape = "array"
	ScalarValue Shape = "scalar"
)

// Value describes logical input/output traffic, not actual memory transactions.
type Value struct {
	Name, TypeID string
	Shape        Shape
	ElementBytes int64
}

// Parameter describes a scalar preparation/identity value, excluded from array
// traffic. The input configuration retains its exact IEEE-754 bits.
type Parameter struct {
	Name, TypeID, Encoding string
}

type WorkerPolicy string

const (
	Serial            WorkerPolicy = "serial"
	PerCallGOMAXPROCS WorkerPolicy = "per-call-gomaxprocs"
)

type Variant struct {
	ID, Label string
	Order     int
	UsesSIMD  bool
	Workers   WorkerPolicy
}

// WorkerCount describes the current kernel policy. A nonpositive procs value
// means unavailable legacy metadata; it is not an assumed runtime setting.
func (v Variant) WorkerCount(elements, procs int) (int, bool) {
	if elements < 0 {
		return 0, false
	}
	switch v.Workers {
	case Serial:
		return 1, true
	case PerCallGOMAXPROCS:
		if elements == 0 {
			return 0, true
		}
		if procs > 0 {
			return min(elements, procs), true
		}
	}
	return 0, false
}

type Operation struct {
	NativePackage                              string // Explicit developer benchmark package, not a runtime dependency.
	ID, Group, Title, BenchmarkName            string
	Category, TypeID, ArithmeticID, Expression string
	ReferenceVariantID, ExecutionMode          string
	Inputs                                     []Value
	Parameters                                 []Parameter
	Output                                     Value
	Variants                                   []Variant
}

// LogicalBytes derives effective work accounting from the declared values.
// Scalar outputs count once; array values count once per element. Constant
// preparation and intermediate traffic are not represented as measured DRAM.
func (op Operation) LogicalBytes(elements int) (int64, error) {
	if elements < 0 {
		return 0, fmt.Errorf("elements must be nonnegative")
	}
	total := int64(0)
	values := append(slices.Clone(op.Inputs), op.Output)
	for _, value := range values {
		if value.TypeID == "" || value.ElementBytes <= 0 {
			return 0, fmt.Errorf("invalid value type for %q", value.Name)
		}
		var count int64
		switch value.Shape {
		case Array:
			count = int64(elements)
		case ScalarValue:
			count = 1
		default:
			return 0, fmt.Errorf("unsupported shape %q", value.Shape)
		}
		if count > (math.MaxInt64-total)/value.ElementBytes {
			return 0, fmt.Errorf("logical byte count overflows int64")
		}
		total += count * value.ElementBytes
	}
	return total, nil
}

var operations = append([]Operation{
	{
		ID: Add, Group: "addition", Title: "Addition", BenchmarkName: "BenchmarkAddition",
		Category: "arithmetic", TypeID: Float32, ArithmeticID: "go-float32-add-v1",
		Expression: "a + b", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"a", Float32, Array, 4}, {"b", Float32, Array, 4}},
		Output: Value{"out", Float32, Array, 4},
		Variants: []Variant{
			{Scalar, "Scalar", 0, false, Serial},
			{SIMDPartial, "SIMDPartial", 2, true, Serial},
			{SIMDScalarTail, "SIMDTail", 3, true, Serial},
			{SIMDParallelScalarTail, "SIMDParallel", 4, true, PerCallGOMAXPROCS},
		},
	},
	{
		ID: ComputeExpression, Group: "computation", Title: "Compute-heavy expression: a*a + b*b + a*b",
		BenchmarkName: "BenchmarkComputation", Category: "composed-expression", TypeID: Float32,
		ArithmeticID: "go-float32-quadratic-v1", Expression: "a*a + b*b + a*b",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs:   []Value{{"a", Float32, Array, 4}, {"b", Float32, Array, 4}},
		Output:   Value{"out", Float32, Array, 4},
		Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMD", 1, true, Serial}},
	},
}, patternOperations()...)

// All returns independent descriptors in the frozen legacy group/variant order.
func All() []Operation {
	list := make([]Operation, len(operations))
	for i, op := range operations {
		list[i] = clone(op)
	}
	return list
}

func Lookup(id, typeID string) (Operation, bool) {
	for _, op := range operations {
		if op.ID == id && op.TypeID == typeID {
			return clone(op), true
		}
	}
	return Operation{}, false
}

func ByGroup(group string) (Operation, bool) {
	for _, op := range operations {
		if op.Group == group {
			return clone(op), true
		}
	}
	return Operation{}, false
}

func ByBenchmark(name string) (Operation, bool) {
	for _, op := range operations {
		if op.BenchmarkName == name {
			return clone(op), true
		}
	}
	return Operation{}, false
}

// LegacyVariant preserves known historical display aliases even when they are
// encountered on an unknown benchmark. It does not register executable cases.
func LegacyVariant(label string) (Variant, bool) {
	for _, op := range operations {
		for _, variant := range op.Variants {
			if variant.Label == label {
				return variant, true
			}
		}
	}
	return Variant{}, false
}

func LegacyLabelOrder(label string) int {
	if variant, ok := LegacyVariant(label); ok {
		return variant.Order
	}
	last := 0
	for _, op := range operations {
		for _, variant := range op.Variants {
			last = max(last, variant.Order)
		}
	}
	return last + 1
}

func clone(op Operation) Operation {
	op.Inputs = slices.Clone(op.Inputs)
	op.Parameters = slices.Clone(op.Parameters)
	op.Variants = slices.Clone(op.Variants)
	return op
}
