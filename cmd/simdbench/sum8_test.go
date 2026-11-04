package main

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/reduction"
	"testing"
)

func TestQuickSum8ProjectedOutput(t *testing.T) {
	o := input.DefaultOptions()
	groups, e := benchmarkGroupsForSize("sum8", 17, nil, nil, nil, o)
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if !g.copyOutput || len(g.output) != 3 || g.scalar != nil || g.logicalBytes != 80 {
		t.Fatal("shape")
	}
	op, _ := catalog.ByGroup("sum8")
	x, _, out := input.MakeForOperation(op, 17, o)
	if e := reduction.Sum8Scalar(x, out); e != nil {
		t.Fatal(e)
	}
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat32(g.output, nil) != input.FingerprintFloat32(out, nil) {
			t.Fatal(e)
		}
	}
}
