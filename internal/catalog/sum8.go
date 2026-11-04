package catalog

const Sum8 = "sum8-f32"
const Blocks8 Shape = "blocks-of-8"

func sum8Operation() Operation {
	return Operation{ID: Sum8, Group: "sum8", Title: "Fixed-eight-value horizontal sum", BenchmarkName: "BenchmarkSum8", NativePackage: "./internal/reduction", Category: "fixed-block-horizontal", TypeID: Float32, ArithmeticID: "float32-sum8-pairs04-leftfold-v1", Expression: "one result per8 inputs: fold-left(x0+x4,x1+x5,x2+x6,x3+x7); pad+0", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Blocks8, 4}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Fixed8-input blocks, pad incomplete block with +0; round four float32 pairs (0+4,1+5,2+6,3+7), then left fold the four pairs. Same order for scalar/SIMD and all native widths. NaN payloads unspecified; signed zeros follow this exact padded order. Output length ceil(n/8), errors before writes, valid empty/nil, disjoint/same-start output-prefix alias; other overlap unsupported. Validation, block padding/copying, loads, pair addition, stores and final fold timed. Logical bytes4*n+4*ceil(n/8), full reduced-array digest/extent after timing. Composed portable algorithm, not isolated ISA horizontal primitive."}
}
