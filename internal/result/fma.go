package result

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/fma32"
)

// FMAIdentity separates timed implementation from the process vector capability.
// VectorPathID/native width must agree across a comparison, including scalar.
type FMAIdentity struct {
	RoundingID   string
	PathID       string
	VectorPathID string
	NativeLanes  OptionalInt
}

func NewFMAIdentity(strategy, vectorPath string, lanes int) (FMAIdentity, error) {
	if vectorPath != fma32.SoftwarePath && vectorPath != fma32.VectorPath {
		return FMAIdentity{}, fmt.Errorf("unknown FMA32 vector path %q", vectorPath)
	}
	if lanes < 1 {
		return FMAIdentity{}, fmt.Errorf("FMA32 requires positive process lanes")
	}
	identity := FMAIdentity{RoundingID: fma32.RoundingID, VectorPathID: vectorPath, PathID: fma32.SoftwarePath, NativeLanes: OptionalInt{Value: lanes, Known: true}}
	switch strategy {
	case catalog.Scalar:
	case catalog.SIMDPartial:
		identity.PathID = vectorPath
	default:
		return FMAIdentity{}, fmt.Errorf("unknown FMA32 strategy %q", strategy)
	}
	return identity, nil
}
