package movement_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func BenchmarkCopy(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Copy) }
