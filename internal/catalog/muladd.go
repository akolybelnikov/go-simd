package catalog

const MulAdd = "muladd-f32"

func muladdOperation() Operation {
	return Operation{ID: MulAdd, Group: "muladd", Title: "Three-array fused multiply-add", BenchmarkName: "BenchmarkMulAdd", NativePackage: "./internal/patterns", Category: "fused-elementwise", TypeID: Float32,
		ArithmeticID: "float32-muladd-fused-rne-v1", Expression: "round32(a*b+c), one rounding", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"a", Float32, Array, 4}, {"b", Float32, Array, 4}, {"c", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4},
		Variants:        []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}},
		NumericalPolicy: "IEEE binary32 round-to-nearest ties-to-even once, including subnormals, overflow and signed zero; NaN payloads unspecified. Scalar uses corrected software FMA32; SIMDPartial uses verified vector MulAdd or recorded software fallback. Three streamed arrays, no broadcasts/reduction. Equal lengths required before writes; valid empty/nil succeeds. Disjoint/exact aliases supported; partial overlap unsupported. Validation, dispatch, loads, FMA and stores timed. Logical bytes 16*n; complete output digest after timing."}
}
