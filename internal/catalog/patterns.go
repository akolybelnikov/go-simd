package catalog

const (
	Scale         = "scale-f32"
	Affine        = "affine-f32"
	ThresholdZero = "threshold-zero-f32"
	SelectGT      = "select-gt-f32"
	RangeZero     = "range-zero-f32"
)

func patternOperations() []Operation {
	var operations []Operation
	for _, spec := range []struct {
		id, group, title, benchmark, expression, semantics string
		names                                              []string
		binary, predicate                                  bool
	}{
		{Scale, "scale", "Broadcast scale", "BenchmarkScale", "alpha*x", "float32-scale-v1", []string{"alpha"}, false, false},
		{Affine, "affine", "Broadcast affine (unfused)", "BenchmarkAffine", "float32(alpha*x)+beta", "float32-affine-unfused-v1", []string{"alpha", "beta"}, false, false},
		{ThresholdZero, "threshold", "Threshold to positive zero", "BenchmarkThreshold", "x > threshold ? x : +0", "float32-threshold-ordered-v1", []string{"threshold"}, false, true},
		{SelectGT, "selection", "Conditional selection", "BenchmarkSelection", "x > threshold ? x : y", "float32-select-ordered-bits-v1", []string{"threshold"}, true, true},
		{RangeZero, "rangefilter", "Range mask composition", "BenchmarkRange", "lower < x && x < upper ? x : +0", "float32-range-ordered-v1", []string{"lower", "upper"}, false, true},
	} {
		op := Operation{
			ID: spec.id, Group: spec.group, Title: spec.title, BenchmarkName: spec.benchmark,
			NativePackage: "./internal/patterns", Category: "broadcast", TypeID: Float32,
			ArithmeticID: spec.semantics, Expression: spec.expression,
			ReferenceVariantID: Scalar, ExecutionMode: KernelOnly,
			Inputs: []Value{{"x", Float32, Array, 4}}, Output: Value{"out", Float32, Array, 4},
			Variants: []Variant{{Scalar, "Scalar", 0, false, Serial}, {SIMDPartial, "SIMDPartial", 2, true, Serial}},
		}
		if spec.binary {
			op.Inputs = append(op.Inputs, Value{"y", Float32, Array, 4})
		}
		if spec.predicate {
			op.Category = "predicate-selection"
		}
		for _, name := range spec.names {
			op.Parameters = append(op.Parameters, Parameter{name, Float32, "ieee754-bits"})
		}
		operations = append(operations, op)
	}
	return operations
}

func Groups() []string {
	var groups []string
	for _, op := range All() {
		groups = append(groups, op.Group)
	}
	return groups
}

func (op Operation) Parameterized() bool { return len(op.Parameters) > 0 }

func (op Operation) NativePath() string {
	if op.NativePackage != "" {
		return op.NativePackage
	}
	return "./internal/" + op.Group
}
