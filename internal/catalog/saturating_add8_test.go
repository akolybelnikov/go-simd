package catalog

import "testing"

func TestSaturatingAdd8Descriptor(t *testing.T) {
	op, ok := ByGroup("saturatingadd8")
	if !ok || op.ID != SaturatingAdd8 || op.TypeID != Uint8 || !op.ByteResult() || op.Output.Shape != Array || len(op.Inputs) != 2 || len(op.Variants) != 2 || op.Parameterized() {
		t.Fatal(op)
	}
	if bytes, e := op.LogicalBytes(1025); e != nil || bytes != 3075 {
		t.Fatal(bytes, e)
	}
}
