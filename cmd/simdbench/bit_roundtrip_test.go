package main

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickBitRoundTrip(t *testing.T) {
	o := input.DefaultOptions()
	groups, e := benchmarkGroupsForSize("bitroundtrip", 17, nil, nil, nil, o)
	if e != nil || len(groups) != 1 {
		t.Fatal(e)
	}
	g := groups[0]
	if !g.copyOutput || g.output == nil || g.byteOutput != nil || g.logicalBytes != 136 {
		t.Fatal("shape")
	}
	op, _ := catalog.ByGroup("bitroundtrip")
	want := input.FingerprintForOperation(op, 17, o)
	for _, v := range g.variants {
		if e := v.run(); e != nil || input.FingerprintFloat32(g.output, nil) != want {
			t.Fatal(e)
		}
	}
}
