package catalog

const Clamp = "clamp-f32"
const SIMDMinMax = "simd-minmax"

func clampOperation() Operation {
	return Operation{
		ID: Clamp, Group: "clamp", Title: "Ordered clamp (bit-preserving ties and NaNs)", BenchmarkName: "BenchmarkClamp",
		NativePackage: "./internal/patterns", Category: "clipping", TypeID: Float32, ArithmeticID: "float32-clamp-ordered-bits-v1",
		Expression:         "x < lower ? lower : x > upper ? upper : x",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4},
		Parameters:      []Parameter{{"lower", Float32, "ieee754-bits"}, {"upper", Float32, "ieee754-bits"}},
		Variants:        []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}, {SIMDMinMax, "SIMDMinMax", 3, true, Serial}},
		NumericalPolicy: "Ordered clipping to exact lower/upper bits. Equality, signed-zero ties and input NaN payloads preserve original bits. Bounds must be non-NaN with lower<=upper; infinities and equal bounds accepted. Length errors take precedence; all failures precede writes, including empty invalid bounds. Disjoint/exact alias; partial overlap unsupported. SIMDMinMax includes explicit outside/tie/NaN repairs; no raw Min/Max equivalence assumed. Validation, broadcasts, selection/repairs and stores timed. Logical bytes 8*n.",
	}
}
