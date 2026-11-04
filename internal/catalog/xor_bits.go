package catalog

const XorBits = "xor-bits-f32"

func xorBitsOperation() Operation {
	return Operation{ID: XorBits, Group: "xorbits", Title: "XOR of opaque Float32 bit patterns", BenchmarkName: "BenchmarkXorBits", NativePackage: "./internal/patterns", Category: "bitwise-representation", TypeID: Float32, ArithmeticID: "float32-opaque-uint32-xor-v1", Expression: "outbits = bits(a) XOR bits(b)", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"a", Float32, Array, 4}, {"b", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Exact uint32 XOR of opaque float32 representations, including NaN payload/sign/exponent bits; no floating arithmetic or numeric conversion. All lengths equal before writes; valid empty/nil, disjoint/exact aliases; partial overlap unsupported. Validation, width setup, bounded loads, bitwise path and stores timed;12*n logical bytes. Complete input/output bit digests after timing; actual Float32/Uint32 lane width retained. Source labels do not prove instruction selection."}
}
