package catalog

import "testing"

func TestSumScalarOutputContract(t *testing.T) {
	op, ok := Lookup("sum-f32", "float32")
	if !ok || op.Group != "sum" || op.BenchmarkName != "BenchmarkSum" || op.ArithmeticID != "float32-sum-ordered-folds-v1" || !op.ScalarResult() || len(op.Inputs) != 1 || len(op.Parameters) != 0 || len(op.Variants) != 4 {
		t.Fatalf("sum descriptor drift: %+v", op)
	}
	for _, n := range []int{0, 1, 1024} {
		got, err := op.LogicalBytes(n)
		if err != nil || got != int64(4*n+4) {
			t.Fatal("scalar-output work accounting lost")
		}
	}
	if op.Variants[0].ID != "scalar" || op.Variants[1].ID != "simd-striped" || op.Variants[1].Label != "SIMDStriped" || op.NumericalPolicy == "" {
		t.Fatal("sum fold identities missing")
	}
}
