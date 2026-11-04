package result

import "fmt"

// VectorIdentity records actual element lanes for typed array experiments.
// Case.TypeID defines the element representation; old experiments retain their
// existing dedicated policy/width fields and timing identities.
type VectorIdentity struct {
	NativeLanes           OptionalInt
	CarrylessCapabilityID string `json:",omitempty"`
}

func NewVectorIdentity(lanes int) (VectorIdentity, error) {
	if lanes < 1 {
		return VectorIdentity{}, fmt.Errorf("vector lanes must be positive")
	}
	return VectorIdentity{NativeLanes: OptionalInt{lanes, true}}, nil
}

func CarrylessCapability(hardware bool) string {
	if hardware {
		return "sdk-hardware-clmul-available/v1"
	}
	return "sdk-software-clmul-only/v1"
}
