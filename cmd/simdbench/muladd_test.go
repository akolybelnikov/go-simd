package main

import (
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickMulAddFullArray(t *testing.T) {
	o := input.DefaultOptions()
	o.FMAInput = "rounding"
	groups, e := benchmarkGroupsForSize("muladd", 17, nil, nil, nil, o)
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if !g.copyOutput || g.logicalBytes != 272 || len(g.variants) != 2 {
		t.Fatal("ternary quick shape")
	}
	a, b, c, out := input.MakeMulAdd(17, o)
	for i := range out {
		out[i] = fma32.FMA(a[i], b[i], c[i])
	}
	want := input.FingerprintFloat32(out, nil)
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat32(g.output, nil) != want {
			t.Fatal(v, e)
		}
	}
}
