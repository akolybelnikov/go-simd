package main

import (
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickSaturatingAdd8TrueOperands(t *testing.T) {
	groups, e := benchmarkGroupsForSize("saturatingadd8", 33, nil, nil, nil, input.DefaultOptions())
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if g.typeID != "uint8" || len(g.byteOutput) != 33 || g.output != nil || g.float64Output != nil || g.logicalBytes != 99 || len(g.variants) != 2 {
		t.Fatal("shape")
	}
	a, b, out := input.MakeSaturatingAdd8(33)
	for i := range out {
		out[i] = uint8(min(int(a[i])+int(b[i]), 255))
	}
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintUint8(g.byteOutput, nil) != input.FingerprintUint8(out, nil) {
			t.Fatal(e)
		}
	}
}
