package result

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
)

const FusedDotRoundingID = "binary32-fma-updates-add-merge-rne/v1"
const SoftwareStripedFMAPath = "software-striped-fma32/v1"
const VectorTailFMAPath = "vector-fma32-with-software-tail/v1"
const SoftwareTailFMAPath = "software-tail-fma32/v1"

// NewDotFMAIdentity records the actual case path, including sizes with no full
// vectors. The process vector gate stays separate from strategy/case execution.
func NewDotFMAIdentity(strategy, vectorPath string, lanes, elements int) (FMAIdentity, error) {
	if elements < 0 {
		return FMAIdentity{}, fmt.Errorf("negative fused-dot size")
	}
	mapped := catalog.Scalar
	switch strategy {
	case catalog.Scalar:
	case catalog.SIMDStriped, catalog.SIMDStriped2, catalog.SIMDStriped4:
		mapped = catalog.SIMDPartial
	default:
		return FMAIdentity{}, fmt.Errorf("unknown fused-dot strategy %q", strategy)
	}
	i, err := NewFMAIdentity(mapped, vectorPath, lanes)
	if err != nil {
		return FMAIdentity{}, err
	}
	i.RoundingID = FusedDotRoundingID
	if strategy != catalog.Scalar {
		if vectorPath == fma32.SoftwarePath {
			i.PathID = SoftwareStripedFMAPath
		} else if elements < lanes {
			i.PathID = SoftwareTailFMAPath
		} else if elements%lanes != 0 {
			i.PathID = VectorTailFMAPath
		}
	}
	return i, nil
}

// DotFMAPathCode is the stable native metric encoding for fused-dot paths only.
func DotFMAPathCode(path string) (int, error) {
	switch path {
	case fma32.SoftwarePath:
		return 0, nil
	case fma32.VectorPath:
		return 1, nil
	case SoftwareStripedFMAPath:
		return 2, nil
	case VectorTailFMAPath:
		return 3, nil
	case SoftwareTailFMAPath:
		return 4, nil
	}
	return 0, fmt.Errorf("unknown fused-dot path %q", path)
}
