package catalog

import "testing"

func TestClmul64Descriptor(t *testing.T) {
	op, ok := ByGroup("clmul64")
	if !ok || op.ID != Clmul64 || op.TypeID != Uint64 || op.ByteResult() || op.Output.Shape != Array || len(op.Inputs) != 2 || len(op.Variants) != 2 || op.Parameterized() {
		t.Fatal(op)
	}
	if bytes, e := op.LogicalBytes(1025); e != nil || bytes != 24600 {
		t.Fatal(bytes, e)
	}
}
