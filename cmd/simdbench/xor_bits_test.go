package main

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"math"
	"testing"
)

func TestQuickXorBits(t *testing.T) {
	groups, e := benchmarkGroupsForSize("xorbits", 17, nil, nil, nil, input.DefaultOptions())
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if !g.copyOutput || g.output == nil || g.logicalBytes != 204 {
		t.Fatal("shape")
	}
	a, b, out := input.MakeXorBits(17)
	for i := range out {
		out[i] = math.Float32frombits(math.Float32bits(a[i]) ^ math.Float32bits(b[i]))
	}
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat32(g.output, nil) != input.FingerprintFloat32(out, nil) {
			t.Fatal(e)
		}
	}
}
