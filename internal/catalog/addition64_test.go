package catalog

import "testing"

func TestAddition64Descriptor(t *testing.T) {
	op, ok := ByGroup("addition64")
	if !ok || op.ID != Add64 || op.TypeID != Float64 || len(op.Inputs) != 2 || len(op.Variants) != 2 || op.Parameterized() || op.Output.ElementBytes != 8 || op.NativePath() != "./internal/addition" {
		t.Fatal(op)
	}
	if bytes, e := op.LogicalBytes(1025); e != nil || bytes != 24600 {
		t.Fatal(bytes, e)
	}
}
