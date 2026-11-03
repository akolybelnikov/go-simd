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

func HarnessFor(op catalog.Operation) string {
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
