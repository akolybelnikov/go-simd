package catalog

const SaturatingAdd8 = "add-sat-u8"

func saturatingAdd8Operation() Operation {
	return Operation{ID: SaturatingAdd8, Group: "saturatingadd8", Title: "Uint8 saturating addition", BenchmarkName: "BenchmarkSaturatingAdd8", NativePackage: "./internal/integer", Category: "integer-saturation", TypeID: Uint8, ArithmeticID: "uint8-add-saturate-255-v1", Expression: "min(uint16(a)+uint16(b),255)", ReferenceVariantID: Scalar, ExecutionMode: KernelOnly, Inputs: []Value{{"a", Uint8, Array, 1}, {"b", Uint8, Array, 1}}, Output: Value{"out", Uint8, Array, 1}, Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}}, NumericalPolicy: "Unsigned byte saturating sum, not wrapping. True uint8 arrays; no float32 conversion or predicate encoding. Equal lengths before writes; valid empty/nil, disjoint/exact aliases, partial overlap unsupported. Validation, width setup, bounded loads, AddSaturated and stores timed; 3*n logical bytes. Full output byte digest and actual Uint8 lanes retained after timing."}
}
