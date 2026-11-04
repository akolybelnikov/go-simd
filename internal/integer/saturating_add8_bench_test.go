package integer_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkSaturatingAdd8(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.SaturatingAdd8) }
