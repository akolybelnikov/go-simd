package catalog

const Copy = "copy-f32"
const GoCopy = "go-copy"

func copyOperation() Operation {
	return Operation{
		ID: Copy, Group: "copy", Title: "Float32 copy (exact bits)",
		BenchmarkName: "BenchmarkCopy", NativePackage: "./internal/movement",
		Category: "data-movement", TypeID: Float32, ArithmeticID: "float32-copy-bits-v1",
		Expression:         "out[i] = x[i], exact float32 bits",
		ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
		Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4},
		Variants:        []Variant{{Scalar, "Scalar", 0, false, Serial}, {GoCopy, "GoCopy", 1, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}, {SIMDScalarTail, "SIMDTail", 3, true, Serial}},
		NumericalPolicy: "Exact input bits, including signed zero, subnormals, infinities and NaN payloads. Equal lengths required; mismatch leaves output unchanged. Empty succeeds. Disjoint storage or exact alias; partial overlap unsupported by the shared contract. Benchmarks use disjoint arrays and include validation, copying and tails. Scalar is a source-loop label; GoCopy uses compiler/runtime-selected Go copy. Logical bytes 8*n, not measured DRAM traffic.",
	}
}
