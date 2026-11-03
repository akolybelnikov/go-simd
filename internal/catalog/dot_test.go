package catalog

import "testing"

func TestDotScalarOutputContract(t *testing.T) {
	op, ok := Lookup("dot-f32", "float32")
	if !ok || op.Group != "dot" || op.BenchmarkName != "BenchmarkDot" || op.NativePath() != "./internal/reduction" || op.ArithmeticID != "float32-dot-unfused-ordered-folds-v1" || !op.ScalarResult() || len(op.Inputs) != 2 || len(op.Parameters) != 0 || len(op.Variants) != 4 {
		t.Fatalf("dot descriptor drift: %+v", op)
	}
	for _, n := range []int{0, 1, 1024} {
		got, err := op.LogicalBytes(n)
		if err != nil || got != int64(8*n+4) {
			t.Fatal("dot scalar-output work accounting lost")
		}
	}
	if op.Variants[0].ID != "scalar" || op.Variants[1].ID != "simd-striped" || op.NumericalPolicy == "" {
		t.Fatal("dot fold identities missing")
	}
}
