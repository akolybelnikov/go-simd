package engine

import (
	"regexp"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

const PatternHarnessID = "standalone-patterns/v1"
const ReductionHarnessID = "standalone-reduction/v1"
const DotHarnessID = "standalone-dot/v1"
const FMAHarnessID = "standalone-affine-fma/v1"
const DotFMAHarnessID = "standalone-dot-fma/v1"
const PredicateHarnessID = "standalone-predicate-export/v1"
const PackedPredicateHarnessID = "standalone-packed-predicate/v1"
const CopyHarnessID = "standalone-copy/v1"
const Sum8HarnessID = "standalone-sum8/v1"
const XorBitsHarnessID = "standalone-xor-bits/v1"
const BitRoundTripHarnessID = "standalone-bit-roundtrip/v1"
const SaturatingAdd8HarnessID = "standalone-saturating-add8/v1"
const Addition64HarnessID = "standalone-addition64/v1"
const MulAddHarnessID = "standalone-muladd-fma/v1"
const ClampHarnessID = "standalone-clamp/v1"

func HarnessFor(op catalog.Operation) string {
	if op.ID == catalog.Sum8 {
		return Sum8HarnessID
	}
	if op.ID == catalog.XorBits {
		return XorBitsHarnessID
	}
	if op.ID == catalog.BitRoundTrip {
		return BitRoundTripHarnessID
	}
	if op.ID == catalog.SaturatingAdd8 {
		return SaturatingAdd8HarnessID
	}
	if op.ID == catalog.Add64 {
		return Addition64HarnessID
	}
	if op.ID == catalog.MulAdd {
		return MulAddHarnessID
	}
	if op.ID == catalog.Clamp {
		return ClampHarnessID
	}
	if op.ID == catalog.Copy {
		return CopyHarnessID
	}
	if op.ID == catalog.PredicateGTPacked {
		return PackedPredicateHarnessID
	}
	if op.ByteResult() {
		return PredicateHarnessID
	}
	if op.ID == catalog.DotFMA {
		return DotFMAHarnessID
	}
	if op.ID == catalog.AffineFMA {
		return FMAHarnessID
	}
	if op.ID == catalog.Dot {
		return DotHarnessID
	}
	if op.ScalarResult() {
		return ReductionHarnessID
	}
	if op.Parameterized() {
		return PatternHarnessID
	}
	return HarnessID
}

// A native package may contain multiple operations. Restrict it to the exact
// prevalidated leaf set so a group never benchmarks sibling operations.
func (p Plan) NativeSelection() string {
	var expressions []string
	for _, c := range p.Cases {
		parts := strings.Split(c.Name, "/")
		for i, s := range parts {
			parts[i] = "^" + regexp.QuoteMeta(s) + "$"
		}
		expressions = append(expressions, strings.Join(parts, "/"))
	}
	if len(expressions) == 0 {
		return "^$"
	}
	return strings.Join(expressions, "|")
}
