package main

import (
	"io"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/input"
)

func TestQuickDotTypedCallbacks(t *testing.T) {
	opts, err := parseOptions([]string{"-group=dot", "-dot-input=cancellation"}, io.Discard)
	if err != nil || opts.input.DotInput != "cancellation" {
		t.Fatal("dot selection failed", err)
	}
	groups, err := benchmarkGroupsForSize("dot", 64, nil, nil, nil, opts.input)
	if err != nil || len(groups) != 1 || groups[0].output != nil || groups[0].scalar == nil || groups[0].logicalBytes != 516 {
		t.Fatal("dot quick preparation failed", err)
	}
	for _, v := range groups[0].variants {
		if v.run != nil || v.reduce != nil || v.reduceChecked == nil {
			t.Fatal("wrong dot callback")
		}
		got, err := v.reduceChecked()
		want := float32(0)
		if v.name != "Scalar" {
			want = 1
		}
		if err != nil || got != want {
			t.Fatal("wrong quick dot result", err, got)
		}
	}
	o := input.DefaultOptions()
	o.DotInput = "unknown"
	if _, err := benchmarkGroupsForSize("dot", 64, nil, nil, nil, o); err == nil {
		t.Fatal("invalid dot fixture accepted")
	}
}
