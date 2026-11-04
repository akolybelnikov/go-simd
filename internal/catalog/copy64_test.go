package catalog

import (
	"math"
	"testing"
)

func TestCopy64UnaryContract(t *testing.T) {
	op, ok := ByGroup("copy64")
	if !ok || op.ID != Copy64 || op.TypeID != Float64 || len(op.Inputs) != 1 || len(op.Variants) != 3 || op.ByteResult() || op.ScalarResult() || op.Parameterized() {
		t.Fatal(op)
	}
	if b, e := op.LogicalBytes(1025); e != nil || b != 16400 {
		t.Fatal(b, e)
	}
	if _, e := op.LogicalBytes(math.MaxInt); e == nil {
		t.Fatal("overflow")
	}
}
