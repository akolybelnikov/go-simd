package catalog

const Dot = "dot-f32"

func dotOperation() Operation {
	return Operation{
		ID: Dot, Group: "dot", Title: "Float32 dot product (unfused, specified fold orders)", BenchmarkName: "BenchmarkDot",
		NativePackage: "./internal/reduction", Category: "reduction", TypeID: Float32,
		ArithmeticID: "float32-dot-unfused-ordered-folds-v1", Expression: "sum(float32(x*y)), float32 rounding per multiply and add",
		NumericalPolicy:    "Products round to float32 before addition, without FMA. Scalar: left-to-right from +0. SIMDStriped: one native-width vector accumulator, low-to-high lane fold from +0, then scalar products in the tail. Cross-strategy bit equality is not required. Empty input returns +0; unequal lengths return ErrLengthMismatch before reading inputs; NaNs/Inf follow the selected fold.",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}, {"y", Float32, Array, 4}}, Output: Value{"dot", Float32, ScalarValue, 4},
		Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDStriped, "SIMDStriped", 1, true, Serial}},
	}
}
