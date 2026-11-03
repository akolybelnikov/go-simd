package reduction_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkDotFMA(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.DotFMA) }
