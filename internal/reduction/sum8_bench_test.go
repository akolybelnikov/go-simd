package reduction_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkSum8(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Sum8) }
