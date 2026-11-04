package patterns_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkMulAdd(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.MulAdd) }
