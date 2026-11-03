package catalog

import "testing"

func TestPredicateDescriptor(t *testing.T) {
	op, ok := ByGroup("predicate")
	if !ok || op.ID != PredicateGT || op.TypeID != Float32 || !op.ByteResult() || op.ScalarResult() || op.Output.TypeID != Uint8 || op.Output.ElementBytes != 1 || len(op.Inputs) != 1 || len(op.Parameters) != 1 || len(op.Variants) != 2 || op.NativePath() != "./internal/patterns" {
		t.Fatal(op)
	}
	for _, n := range []int{0, 1, 9, 1024} {
		if b, e := op.LogicalBytes(n); e != nil || b != int64(5*n) {
			t.Fatal(b, e)
		}
	}
}
