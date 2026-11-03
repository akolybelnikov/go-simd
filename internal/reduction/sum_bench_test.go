package reduction_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkSum(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Sum) }
