package catalog

import "testing"

func TestFusedDotDescriptor(t *testing.T) {
	op, ok := ByGroup("dotfma")
	if !ok || op.ID != "dot-fma-f32" || op.ArithmeticID != "float32-dot-fused-ordered-folds-v1" || op.BenchmarkName != "BenchmarkDotFMA" || op.NativePath() != "./internal/reduction" || !op.ScalarResult() || len(op.Inputs) != 2 || len(op.Parameters) != 0 || len(op.Variants) != 4 {
		t.Fatal("fused-dot descriptor drift", op)
	}
	for _, n := range []int{0, 1, 64, 1024} {
		got, err := op.LogicalBytes(n)
		if err != nil || got != int64(8*n+4) {
			t.Fatal("fused-dot byte model drift")
		}
	}
}
