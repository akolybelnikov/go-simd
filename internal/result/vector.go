package result

import "fmt"

// VectorIdentity records actual element lanes for typed array experiments.
// Case.TypeID defines the element representation; old experiments retain their
// existing dedicated policy/width fields and timing identities.
type VectorIdentity struct{ NativeLanes OptionalInt }

func NewVectorIdentity(lanes int) (VectorIdentity, error) {
	if lanes < 1 {
		return VectorIdentity{}, fmt.Errorf("vector lanes must be positive")
	}
	return VectorIdentity{OptionalInt{lanes, true}}, nil
}
