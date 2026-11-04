package main

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/integer"
	"testing"
)

func TestQuickClmul64TrueWords(t *testing.T) {
	gs, e := benchmarkGroupsForSize("clmul64", 17, nil, nil, nil, input.DefaultOptions())
	if e != nil || len(gs) != 1 {
		t.Fatal(e)
	}
	g := gs[0]
	if g.output != nil || g.float64Output != nil || len(g.uint64Output) != 17 || g.logicalBytes != 408 {
		t.Fatal("typed shape")
	}
	a, b, out := input.MakeClmul64(17)
	if e := integer.Clmul64Scalar(a, b, out); e != nil {
		t.Fatal(e)
	}
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintUint64(g.uint64Output, nil) != input.FingerprintUint64(out, nil) {
			t.Fatal(e)
		}
	}
}
