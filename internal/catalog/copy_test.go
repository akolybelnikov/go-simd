package catalog

import (
	"math"
	"slices"
	"testing"
)

func TestCopyDescriptor(t *testing.T) {
	op, ok := ByGroup("copy")
	if !ok || op.ID != Copy || op.TypeID != Float32 || op.Output.TypeID != Float32 || op.Parameterized() || op.ScalarResult() || op.ByteResult() || len(op.Inputs) != 1 || op.NativePath() != "./internal/movement" {
		t.Fatal(op)
	}
	want := []string{"Scalar", "GoCopy", "SIMDPartial", "SIMDTail"}
	var labels []string
	for _, v := range op.Variants {
		labels = append(labels, v.Label)
		if v.Workers != Serial {
			t.Fatal(v)
		}
	}
	if !slices.Equal(labels, want) || op.Variants[1].ID != GoCopy || op.Variants[1].UsesSIMD {
		t.Fatal(labels)
	}
	for _, n := range []int{0, 1, 9, 1024, 1 << 24} {
		if b, e := op.LogicalBytes(n); e != nil || b != int64(8*n) {
			t.Fatal(n, b, e)
		}
	}
	if _, e := op.LogicalBytes(math.MaxInt); e == nil {
		t.Fatal("copy work overflow")
	}
}
