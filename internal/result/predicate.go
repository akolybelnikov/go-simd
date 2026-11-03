package result

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

const PredicateEncodingID = "uint8-zero-one/v1"
const PackedPredicateEncodingID = "uint8-packed-lsb0/v1"

func NewPackedPredicateIdentity(lanes int) (PredicateIdentity, error) {
	i, err := NewPredicateIdentity(lanes)
	if err != nil {
		return PredicateIdentity{}, err
	}
	i.EncodingID = PackedPredicateEncodingID
	return i, nil
}

type PredicateIdentity struct {
	EncodingID  string
	NativeLanes OptionalInt
}

func NewPredicateIdentity(lanes int) (PredicateIdentity, error) {
	if lanes < 1 {
		return PredicateIdentity{}, fmt.Errorf("predicate width must be positive")
	}
	return PredicateIdentity{PredicateEncodingID, OptionalInt{lanes, true}}, nil
}

// Eight exact uint32 metrics encode the SHA-256 digest in big-endian word order.
// This pure importer never substitutes expected/parent output for child facts.
func predicateDigest(metrics map[string]Metric) (string, error) {
	var digest [32]byte
	for i := 0; i < 8; i++ {
		m := metrics[fmt.Sprintf("output-sha256-%d", i)]
		if !m.Available || !finite(m.Value) || m.Value < 0 || m.Value > math.MaxUint32 || math.Trunc(m.Value) != m.Value {
			return "", fmt.Errorf("predicate requires eight exact uint32 output-sha256 metrics")
		}
		binary.BigEndian.PutUint32(digest[i*4:], uint32(m.Value))
	}
	return hex.EncodeToString(digest[:]), nil
}
