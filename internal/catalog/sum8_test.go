package catalog

import (
	"math"
	"testing"
)

func TestSum8ShapeExtentAndTraffic(t *testing.T) {
	op, ok := ByGroup("sum8")
	if !ok || op.ID != Sum8 || op.Output.Shape != Blocks8 || op.ScalarResult() || op.ByteResult() || op.Parameterized() || len(op.Inputs) != 1 || len(op.Variants) != 2 {
		t.Fatal(op)
	}
	for _, n := range []int{0, 1, 7, 8, 9, 15, 16, 17, 1025} {
		count := n / 8
		if n%8 != 0 {
			count++
		}
		extent, e := op.OutputElements(n)
		if e != nil || extent != count {
			t.Fatal(extent, e)
		}
		bytes, e := op.LogicalBytes(n)
		if e != nil || bytes != int64(4*n+4*count) {
			t.Fatal(bytes, e)
		}
	}
	if _, e := op.OutputElements(-1); e == nil {
		t.Fatal("negative extent")
	}
	if _, e := op.LogicalBytes(math.MaxInt); e == nil {
		t.Fatal("overflow")
	}
}
