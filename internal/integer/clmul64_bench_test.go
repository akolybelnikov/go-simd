package integer_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkClmul64(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Clmul64) }
