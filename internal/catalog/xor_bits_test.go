package catalog

import "testing"

func TestXorBitsDescriptor(t *testing.T) {
	op, ok := ByGroup("xorbits")
	if !ok || op.ID != XorBits || op.TypeID != Float32 || len(op.Inputs) != 2 || len(op.Variants) != 2 || op.Parameterized() || op.Output.Shape != Array {
		t.Fatal(op)
	}
	if bytes, e := op.LogicalBytes(1025); e != nil || bytes != 12300 {
		t.Fatal(bytes, e)
	}
}
