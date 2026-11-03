package catalog

import "testing"

func TestFusedAffineDescriptor(t *testing.T) {
	op, ok := ByGroup("affinefma")
	if !ok || op.ID != "affine-fma-f32" || op.ArithmeticID != "float32-affine-fused-rne-v1" || op.BenchmarkName != "BenchmarkAffineFMA" || op.NativePath() != "./internal/patterns" || len(op.Inputs) != 1 || op.ScalarResult() || len(op.Parameters) != 2 || len(op.Variants) != 2 || op.NumericalPolicy == "" {
		t.Fatal("fused descriptor drift", op)
	}
	bytes, err := op.LogicalBytes(1024)
	if err != nil || bytes != 8192 {
		t.Fatal("FMA broadcasts are not extra input arrays")
	}
	count := 0
	for _, entry := range All() {
		count += len(entry.Variants)
	}
	if count != 32 {
		t.Fatal("implementation inventory", count)
	}
}
