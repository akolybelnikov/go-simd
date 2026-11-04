package main

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickTrueFloat64Output(t *testing.T) {
	groups, e := benchmarkGroupsForSize("addition64", 17, nil, nil, nil, input.DefaultOptions())
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if g.typeID != "float64" || len(g.float64Output) != 17 || g.output != nil || g.logicalBytes != 408 || len(g.variants) != 2 {
		t.Fatal("shape")
	}
	a, b, out := input.MakeAddition64(17)
	for i := range out {
		out[i] = a[i] + b[i]
	}
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat64(g.float64Output, nil) != input.FingerprintFloat64(out, nil) {
			t.Fatal(e)
		}
	}
}
