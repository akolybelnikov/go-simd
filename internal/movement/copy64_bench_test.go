package movement_test

import (
	"github.com/akolybelnikov/go-simd/internal/benchutil"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"testing"
)

func BenchmarkCopy64(b *testing.B) { benchutil.BenchmarkOperation(b, catalog.Copy64) }
