package catalog

import "fmt"

const Uint8 = "uint8"
const PredicateGT = "predicate-gt-f32"

func predicateOperation() Operation {
	return Operation{
		ID: PredicateGT, Group: "predicate", Title: "Ordered predicate export (uint8 0/1)",
		BenchmarkName: "BenchmarkPredicate", NativePackage: "./internal/patterns",
		Category: "predicate-export", TypeID: Float32, ArithmeticID: "float32-predicate-gt-uint8-01-v1",
		Expression:         "out[i] = x[i] > threshold ? uint8(1) : uint8(0)",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Uint8, Array, 1},
		Parameters:      []Parameter{{"threshold", Float32, "ieee754-bits"}},
		Variants:        []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}},
		NumericalPolicy: "Ordered strict comparison; equality or NaN yields byte 0, true yields byte 1. SIMD partial vectors select explicit 0/1, convert to int32 scratch and write valid bytes only. Conversion, extraction and stores are timed. Disjoint storage; unequal lengths fail before writes. Empty succeeds. Logical bytes 5*n; this is not packed bits or raw SDK mask storage.",
	}
}

func (op Operation) ByteResult() bool {
	return (op.Output.Shape == Array || op.Output.Shape == PackedBits) && op.Output.TypeID == Uint8
}

// OutputElements gives physical storage, distinct from logical input size.
// Division before rounding avoids n+7 overflow for packed representations.
func (op Operation) OutputElements(elements int) (int, error) {
	if elements < 0 {
		return 0, fmt.Errorf("elements must be nonnegative")
	}
	switch op.Output.Shape {
	case Array:
		return elements, nil
	case ScalarValue:
		return 1, nil
	case PackedBits, Blocks8:
		n := elements / 8
		if elements%8 != 0 {
			n++
		}
		return n, nil
	}
	return 0, fmt.Errorf("unsupported output shape %q", op.Output.Shape)
}
