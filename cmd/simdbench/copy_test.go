package main

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestQuickUnaryCopyOutput(t *testing.T) {
	o := input.DefaultOptions()
	groups, err := benchmarkGroupsForSize("copy", 17, nil, nil, nil, o)
	if err != nil || len(groups) != 1 {
		t.Fatal(err)
	}
	g := groups[0]
	if !g.copyOutput || len(g.output) != 17 || g.scalar != nil || g.byteOutput != nil || g.logicalBytes != 136 || len(g.variants) != 4 {
		t.Fatal("copy quick shape")
	}
	op, _ := catalog.ByGroup("copy")
	want := input.FingerprintForOperation(op, 17, o)
	for _, v := range g.variants {
		if err := v.run(); err != nil || input.FingerprintFloat32(g.output, nil) != want {
			t.Fatal(err)
		}
	}
}
