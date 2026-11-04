package benchutil

import (
	"slices"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestPackedBoundarySweepAndPreservedSizes(t *testing.T) {
	for _, op := range catalog.All() {
		for _, w := range []int{4, 8, 16} {
			sizes := OperationSizes(op, w)
			if op.TypeID == catalog.Uint8 {
				if !slices.Equal(sizes, BenchmarkSizes(w*4)) {
					t.Fatal("uint8 sweep uses wrong width")
				}
				continue
			}
			if op.TypeID == catalog.Float64 || op.TypeID == catalog.Uint64 {
				if !slices.Equal(sizes, BenchmarkSizes(w/2)) {
					t.Fatal("64-bit sweep uses wrong width")
				}
				continue
			}
			if op.Output.Shape == catalog.Blocks8 {
				for _, n := range []int{7, 8, 9, 15, 16, 17, w - 1, w, w + 1} {
					if !slices.Contains(sizes, n) {
						t.Fatal("missing block/native boundary", w, n)
					}
				}
				continue
			}
			if op.ID != catalog.PredicateGTPacked {
				if !slices.Equal(sizes, BenchmarkSizes(w)) {
					t.Fatal("original sweep changed")
				}
				continue
			}
			for _, n := range []int{1, 7, 8, 9, 63, 64, 65, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1} {
				if !slices.Contains(sizes, n) {
					t.Fatal(w, n)
				}
			}
			if !slices.IsSorted(sizes) || !slices.Equal(sizes, slices.Compact(slices.Clone(sizes))) {
				t.Fatal("duplicate/unordered sizes")
			}
		}
	}
}
