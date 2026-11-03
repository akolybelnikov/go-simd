package catalog

import (
	"math"
	"testing"
)

func TestPackedPredicateDescriptorAndStorage(t *testing.T) {
	op, ok := ByGroup("predicatepack")
	if !ok || op.ID != PredicateGTPacked || op.TypeID != Float32 || !op.ByteResult() || op.ScalarResult() || op.Output.Shape != PackedBits || op.Output.TypeID != Uint8 || len(op.Inputs) != 1 || len(op.Parameters) != 1 || len(op.Variants) != 2 {
		t.Fatal(op)
	}
	for _, tt := range [][2]int{{0, 0}, {1, 1}, {7, 1}, {8, 1}, {9, 2}, {63, 8}, {64, 8}, {65, 9}, {1025, 129}} {
		n, want := tt[0], tt[1]
		if got, err := op.OutputElements(n); err != nil || got != want {
			t.Fatal(n, got, err)
		}
		if b, err := op.LogicalBytes(n); err != nil || b != int64(4*n+want) {
			t.Fatal(n, b, err)
		}
	}
	if n, e := op.OutputElements(math.MaxInt); e != nil || n != math.MaxInt/8+1 {
		t.Fatal("packed count overflow", n, e)
	}
	if _, e := op.LogicalBytes(math.MaxInt); e == nil {
		t.Fatal("traffic overflow accepted")
	}
	if _, e := op.OutputElements(-1); e == nil {
		t.Fatal("negative count")
	}
	op.Output.TypeID = Float32
	if _, e := op.LogicalBytes(8); e == nil {
		t.Fatal("invalid packed type")
	}
	op.Output.Shape = "unknown"
	if _, e := op.OutputElements(8); e == nil {
		t.Fatal("unknown shape")
	}
	plain, _ := ByGroup("predicate")
	if n, e := plain.OutputElements(9); e != nil || n != 9 {
		t.Fatal(n, e)
	}
}
