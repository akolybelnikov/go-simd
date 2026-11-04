package catalog

const BitRoundTrip = "bit-roundtrip-f32"

func bitRoundTripOperation() Operation {
	return Operation{ID: BitRoundTrip, Group: "bitroundtrip", Title: "Float32 exact bit reinterpretation round trip", BenchmarkName: "BenchmarkBitRoundTrip", NativePackage: "./internal/patterns", Category: "bit-representation", TypeID: Float32, ArithmeticID: "float32-uint32-bit-roundtrip-v1", Expression: "Float32frombits(Float32bits(x)), exact representation", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Exact float32 bit preservation through uint32 reinterpretation, including signed zeros, subnormals, infinities and quiet/signaling NaN payloads; no numeric conversion. Length errors before writes, valid empty/nil, disjoint/exact alias; partial overlap unsupported. Validation, width setup, bounded loads, reinterpretation path and stores timed; 8*n logical bytes, full output digest after timing. Compiler may eliminate representational steps; labels/timing do not establish instruction emission."}
}
