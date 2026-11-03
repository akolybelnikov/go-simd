package main

import (
	"bytes"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestQuickPredicateTypedOutput(t *testing.T) {
	o := input.DefaultOptions()
	groups, err := benchmarkGroupsForSize("predicate", 5, nil, nil, nil, o)
	if err != nil || len(groups) != 1 {
		t.Fatal(err)
	}
	g := groups[0]
	if g.output != nil || g.scalar != nil || len(g.byteOutput) != 5 || g.logicalBytes != 25 {
		t.Fatal("wrong quick output shape")
	}
	for _, v := range g.variants {
		if err := v.run(); err != nil || !bytes.Equal(g.byteOutput, []byte{1, 0, 1, 0, 1}) {
			t.Fatal(g.byteOutput, err)
		}
	}
}
