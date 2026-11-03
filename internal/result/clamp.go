package result

import "fmt"

const ClampPolicyID = "ordered-clamp-preserve-bits/v1"

type ClampIdentity struct {
	PolicyID    string
	NativeLanes OptionalInt
}

func NewClampIdentity(lanes int) (ClampIdentity, error) {
	if lanes < 1 {
		return ClampIdentity{}, fmt.Errorf("clamp width must be positive")
	}
	return ClampIdentity{ClampPolicyID, OptionalInt{lanes, true}}, nil
}
