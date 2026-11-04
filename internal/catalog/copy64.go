package catalog

const Copy64 = "copy-f64"

func copy64Operation() Operation {
	return Operation{ID: Copy64, Group: "copy64", Title: "Exact Float64 copy", BenchmarkName: "BenchmarkCopy64", NativePackage: "./internal/movement", Category: "data-movement", TypeID: Float64, ArithmeticID: "binary64-exact-copy-v1", Expression: "out=bits-exact(x), no arithmetic", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"x", Float64, Array, 8}}, Output: Value{"out", Float64, Array, 8}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {GoCopy, "GoCopy", 1, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Exact binary64 bits, including signed zeros, subnormals, infinities and all quiet/signaling NaN payloads. Equal lengths before writes; empty/nil, disjoint/exact aliases valid; partial overlap unsupported. Time validation, width setup, bounded loads/stores or source loop/Go builtin. GoCopy is a compiler/runtime control, unaffected by SDK emulation. 16*n logical bytes, complete output digest and actual Float64 lanes retained after timing."}
}
