package catalog

const AffineFMA = "affine-fma-f32"

func fmaOperation() Operation {
	return Operation{
		ID: AffineFMA, Group: "affinefma", Title: "Broadcast affine (fused binary32)", BenchmarkName: "BenchmarkAffineFMA", NativePackage: "./internal/patterns", Category: "fused-affine", TypeID: Float32,
		ArithmeticID: "float32-affine-fused-rne-v1", Expression: "round32(alpha*x+beta), one rounding",
		NumericalPolicy:    "IEEE binary32 round-to-nearest ties-to-even once; includes subnormals, signed zero and overflow. NaN payloads unspecified. Scalar uses corrected software FMA32; SIMDPartial uses process-verified MulAdd or an explicitly recorded software path. Existing affine and dot remain unfused.",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4}, Parameters: []Parameter{{"alpha", Float32, "ieee754-bits"}, {"beta", Float32, "ieee754-bits"}},
		Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}},
	}
}
