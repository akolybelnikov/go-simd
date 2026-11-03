package result

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

// FoldIdentity is part of the complete measurement key. Order may differ across
// strategies under the declared sum contract; native width must agree across
// an entire comparison. Missing native widths are rejected for new sum text.
type FoldIdentity struct {
	OrderID            string
	NativeLanes        OptionalInt
	VectorAccumulators int
	TailID             string
}

func SumFold(strategy string, lanes int) (FoldIdentity, error) {
	if lanes < 1 {
		return FoldIdentity{}, fmt.Errorf("sum requires a positive native lane count")
	}
	fold := FoldIdentity{NativeLanes: OptionalInt{Value: lanes, Known: true}, TailID: "none"}
	switch strategy {
	case catalog.Scalar:
		fold.OrderID = "left-to-right-f32/v1"
	case catalog.SIMDStriped:
		fold.OrderID = "striped-lane-fold-f32/v1"
		fold.VectorAccumulators = 1
		fold.TailID = "scalar-after-lanes/v1"
	default:
		return FoldIdentity{}, fmt.Errorf("unregistered sum strategy %q", strategy)
	}
	return fold, nil
}
