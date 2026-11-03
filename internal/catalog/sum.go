package catalog

const Sum = "sum-f32"
const SIMDStriped = "simd-striped"

func sumOperation() Operation {
	return Operation{
		ID: Sum, Group: "sum", Title: "Float32 sum (specified fold orders)", BenchmarkName: "BenchmarkSum",
		NativePackage: "./internal/reduction", Category: "reduction", TypeID: Float32,
		ArithmeticID: "float32-sum-ordered-folds-v1", Expression: "sum(x), float32 rounding per addition",
		NumericalPolicy:    "Scalar: left-to-right from +0. SIMDStriped: one native-width vector accumulator, low-to-high lane fold from +0, then scalar tail. Each addition rounds to float32; cross-strategy bit equality is not required. Empty input returns +0; NaNs/Inf follow the selected fold.",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"sum", Float32, ScalarValue, 4},
		Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDStriped, "SIMDStriped", 1, true, Serial}},
	}
}

func (op Operation) ScalarResult() bool { return op.Output.Shape == ScalarValue }
