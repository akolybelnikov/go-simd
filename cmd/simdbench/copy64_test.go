package main

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickCopy64TrueUnary(t *testing.T) {
	gs, e := benchmarkGroupsForSize("copy64", 17, nil, nil, nil, input.DefaultOptions())
	if e != nil || len(gs) != 1 {
		t.Fatal(e)
	}
	g := gs[0]
	if g.output != nil || g.uint64Output != nil || len(g.float64Output) != 17 || g.logicalBytes != 272 || len(g.variants) != 3 {
		t.Fatal("typed unary shape")
	}
	x, _ := input.MakeCopy64(17)
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat64(g.float64Output, nil) != input.FingerprintFloat64(x, nil) {
			t.Fatal(e)
		}
	}
}
