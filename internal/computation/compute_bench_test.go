package computation_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkComputation(b *testing.B) {
	benchutil.BenchmarkOperation(b, catalog.ComputeExpression)
}
