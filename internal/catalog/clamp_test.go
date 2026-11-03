package catalog

import "testing"

func TestClampDescriptor(t *testing.T) {
	op, ok := ByGroup("clamp")
	if !ok || op.ID != Clamp || op.TypeID != Float32 || !op.Parameterized() || op.ScalarResult() || op.ByteResult() || len(op.Inputs) != 1 || len(op.Parameters) != 2 || len(op.Variants) != 3 || op.Variants[2].ID != SIMDMinMax || op.NativePath() != "./internal/patterns" {
		t.Fatal(op)
	}
	for _, n := range []int{0, 1, 9, 1024} {
		if b, e := op.LogicalBytes(n); e != nil || b != int64(8*n) {
			t.Fatal(b, e)
		}
	}
}
