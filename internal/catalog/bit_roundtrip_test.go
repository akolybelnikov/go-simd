package catalog

import "testing"

func TestBitRoundTripDescriptor(t *testing.T) {
	op, ok := ByGroup("bitroundtrip")
	if !ok || op.ID != BitRoundTrip || op.TypeID != Float32 || len(op.Inputs) != 1 || len(op.Variants) != 2 || op.Parameterized() || op.Output.Shape != Array {
		t.Fatal(op)
	}
	if bytes, e := op.LogicalBytes(1025); e != nil || bytes != 8200 {
		t.Fatal(bytes, e)
	}
}
