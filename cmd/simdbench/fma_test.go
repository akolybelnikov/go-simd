package main

import (
	"bytes"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/cli"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/patterns"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

func TestFMAQuickAndHumanListing(t *testing.T) {
	opts, err := parseOptions([]string{"-group=affinefma", "-fma-input=rounding", "-alpha=1.0000001192092896", "-beta=-1"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := benchmarkGroupsForSize("affinefma", 8, nil, nil, nil, opts.input)
	if err != nil || len(groups) != 1 || groups[0].logicalBytes != 64 {
		t.Fatal("FMA quick preparation", err)
	}
	for _, v := range groups[0].variants {
		if err := v.run(); err != nil {
			t.Fatal(err)
		}
		if got := groups[0].output[7]; math.Float32bits(got) != math.Float32bits(-0x1p-46) {
			t.Fatal("quick FMA unexpectedly unfused", got)
		}
	}
	var listing bytes.Buffer
	if err := cli.PrintCatalog(&listing, "affinefma", platform.Current()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listing.String(), patterns.FMAPath()) || !strings.Contains(listing.String(), "software FMA32") {
		t.Fatal("listing hides FMA path", listing.String())
	}
	if patterns.FMAPath() == fma32.VectorPath && !strings.Contains(listing.String(), "verified vector FMA32") {
		t.Fatal("listing hides verified vectors")
	}
}
