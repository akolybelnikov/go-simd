package main

import (
	"bytes"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestQuickPackedPredicateOutputExtent(t *testing.T) {
	groups, err := benchmarkGroupsForSize("predicatepack", 9, nil, nil, nil, input.DefaultOptions())
	if err != nil || len(groups) != 1 {
		t.Fatal(err)
	}
	g := groups[0]
	if g.output != nil || g.scalar != nil || len(g.byteOutput) != 2 || g.logicalBytes != 38 {
		t.Fatal("wrong packed storage")
	}
	for _, v := range g.variants {
		if err := v.run(); err != nil || !bytes.Equal(g.byteOutput, []byte{0x55, 1}) {
			t.Fatal(g.byteOutput, err)
		}
	}
}
