package catalog

const DotFMA = "dot-fma-f32"

func dotFMAOperation() Operation {
	return Operation{
		ID: DotFMA, Group: "dotfma", Title: "Float32 fused dot (specified fold orders)", BenchmarkName: "BenchmarkDotFMA", NativePackage: "./internal/reduction", Category: "fused-reduction", TypeID: Float32,
		ArithmeticID: "float32-dot-fused-ordered-folds-v1", Expression: "FMA32(x,y,accumulator), float32 merge/fold additions",
		NumericalPolicy:    "One binary32 nearest-even rounding per multiply-accumulate; not one rounding for the whole dot. Scalar folds left to right from +0. SIMDStriped/2/4 assign full vectors round-robin to 1/2/4 accumulators, merge lane-wise left to right, fold lanes low to high from +0, then software FMA tail updates. Gated vector FMA or software-striped execution preserves the declared order. Empty returns +0; unequal lengths fail before reading input. NaN payloads unspecified; signed zero/overflow follow each fold. Existing dot remains unfused.",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}, {"y", Float32, Array, 4}}, Output: Value{"dot", Float32, ScalarValue, 4},
		Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDStriped, "SIMDStriped", 1, true, Serial}, {SIMDStriped2, "SIMDStriped2", 2, true, Serial}, {SIMDStriped4, "SIMDStriped4", 3, true, Serial}},
	}
}
