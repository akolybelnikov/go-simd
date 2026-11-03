package result

import "fmt"

const CopyBitPolicyID = "float32-exact-bits/v1"

type CopyIdentity struct {
	BitPolicyID string
	NativeLanes OptionalInt
}

func NewCopyIdentity(lanes int) (CopyIdentity, error) {
	if lanes < 1 {
		return CopyIdentity{}, fmt.Errorf("copy width must be positive")
	}
	return CopyIdentity{CopyBitPolicyID, OptionalInt{lanes, true}}, nil
}
