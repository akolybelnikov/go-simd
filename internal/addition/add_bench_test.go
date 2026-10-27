package addition_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkAddition(b *testing.B) {
	benchutil.BenchmarkOperation(b, catalog.Add)
}
