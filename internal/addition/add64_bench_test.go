package addition_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkAddition64(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Add64) }
