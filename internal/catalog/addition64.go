package catalog

const Float64 = "float64"
const Add64 = "add-f64"

func addition64Operation() Operation {
	return Operation{ID: Add64, Group: "addition64", Title: "Float64 addition", BenchmarkName: "BenchmarkAddition64", NativePackage: "./internal/addition", Category: "addition", TypeID: Float64, ArithmeticID: "go-float64-add-v1", Expression: "a+b (binary64)", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"a", Float64, Array, 8}, {"b", Float64, Array, 8}}, Output: Value{"out", Float64, Array, 8}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}},
		NumericalPolicy: "Ordinary binary64 addition with IEEE signed-zero, subnormal, infinity and overflow behavior; NaN payloads unspecified. True float64 fixtures, no float32 widening. Equal lengths before writes, valid empty/nil, disjoint/exact aliases; partial overlap unsupported. Validation, loads, addition and stores timed; 24*n logical bytes. Complete little-endian float64 output digest after timing; actual float64 lanes retained."}
}
