package main

import (
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/input"
	"testing"
)

func TestQuickClampWholeOutput(t *testing.T) {
	options := input.DefaultOptions()
	options.Pattern = "shuffled"
	options.Seed = 123
	groups, err := benchmarkGroupsForSize("clamp", 17, nil, nil, nil, options)
	if err != nil || len(groups) != 1 {
		t.Fatal(err)
	}
	g := groups[0]
	if !g.copyOutput || len(g.output) != 17 || g.scalar != nil || g.byteOutput != nil || g.logicalBytes != 136 || len(g.variants) != 3 {
		t.Fatal("clamp quick shape")
	}
	op, _ := catalog.ByGroup("clamp")
	x, _, out := input.MakeForOperation(op, 17, options)
	for i, value := range x {
		out[i] = value
		if value < -1 {
			out[i] = -1
		}
		if value > 1 {
			out[i] = 1
		}
	}
	want := input.FingerprintFloat32(out, nil)
	for _, v := range g.variants {
		if err := v.run(); err != nil || input.FingerprintFloat32(g.output, nil) != want {
			t.Fatal(v, err)
		}
	}
}
