package patterns_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkPredicatePacked(b *testing.B) {
	benchutil.BenchmarkOperation(b, catalog.PredicateGTPacked)
}
