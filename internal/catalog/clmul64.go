package catalog

const Uint64 = "uint64"
const Clmul64 = "clmul-low64-u64"

func clmul64Operation() Operation {
	return Operation{ID: Clmul64, Group: "clmul64", Title: "Low64 carryless polynomial multiplication", BenchmarkName: "BenchmarkClmul64", NativePackage: "./internal/integer", Category: "carryless-polynomial", TypeID: Uint64, ArithmeticID: "gf2-polynomial-product-mod-x64-v1", Expression: "low64(a carryless-multiply b)", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"a", Uint64, Array, 8}, {"b", Uint64, Array, 8}}, Output: Value{"out", Uint64, Array, 8}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Exact polynomial multiplication over GF(2) modulo x^64; XOR coefficient addition, high product bits discarded. True uint64 arrays, no floating conversion. Validate lengths and SDK widths 2/4/8 before writes; empty/disjoint/exact aliases valid, partial overlap unsupported. Both Even/Odd products, scratch stores, extraction and output writes timed; 24*n logical bytes. Retain actual lanes, complete output bits and SDK carryless capability; capability is not emitted-instruction proof or a cryptographic claim."}
}
