package catalog

import "testing"

func TestMulAddDescriptor(t *testing.T) {
	op, ok := ByGroup("muladd")
	if !ok || op.ID != MulAdd || op.Parameterized() || op.ScalarResult() || op.ByteResult() || len(op.Inputs) != 3 || len(op.Variants) != 2 || op.NativePath() != "./internal/patterns" || op.ArithmeticID != "float32-muladd-fused-rne-v1" {
		t.Fatal(op)
	}
	if n, e := op.LogicalBytes(1025); e != nil || n != 16400 {
		t.Fatal(n, e)
	}
}
