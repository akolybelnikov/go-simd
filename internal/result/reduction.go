package result

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

// FoldIdentity is part of the complete measurement key. Order may differ across
// strategies under the declared reduction contract; native width must agree
// across a comparison. Missing native widths are rejected for reduction text.
type FoldIdentity struct {
	OrderID            string
	NativeLanes        OptionalInt
	VectorAccumulators int
	TailID             string
}

// ReductionFold preserves sum identities and distinguishes unfused dot orders.
func ReductionFold(operation, strategy string, lanes int) (FoldIdentity, error) {
	if operation == catalog.Sum {
		return SumFold(strategy, lanes)
	}
	if operation != catalog.Dot {
		return FoldIdentity{}, fmt.Errorf("unregistered reduction %q", operation)
	}
	fold, err := SumFold(strategy, lanes)
	if err != nil {
		return FoldIdentity{}, err
	}
	if strategy == catalog.Scalar {
		fold.OrderID = "left-to-right-unfused-dot-f32/v1"
	} else {
		fold.OrderID = "striped-lane-unfused-dot-f32/v1"
		fold.TailID = "scalar-products-after-lanes/v1"
	}
	return fold, nil
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
