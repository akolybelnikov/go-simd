package benchutil

import (
	"slices"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

// OperationSizes preserves the original sweep, adding byte-packing boundaries
// for the packed representation in both compiled and native selection.
func OperationSizes(op catalog.Operation, width int) []int {
	// width is the process float32 lane count; float64 has half as many lanes.
	if op.TypeID == catalog.Float64 {
		width = max(1, width/2)
	}
	sizes := BenchmarkSizes(width)
	if op.ID == catalog.PredicateGTPacked {
		sizes = append(sizes, 7, 9, 63, 65)
		slices.Sort(sizes)
		sizes = slices.Compact(sizes)
	}
	return sizes
}
