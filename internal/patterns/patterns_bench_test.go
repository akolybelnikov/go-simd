package patterns_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkScale(b *testing.B)     { benchutil.BenchmarkOperation(b, catalog.Scale) }
func BenchmarkAffine(b *testing.B)    { benchutil.BenchmarkOperation(b, catalog.Affine) }
func BenchmarkThreshold(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.ThresholdZero) }
func BenchmarkSelection(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.SelectGT) }
func BenchmarkRange(b *testing.B)     { benchutil.BenchmarkOperation(b, catalog.RangeZero) }
