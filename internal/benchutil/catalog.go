package benchutil

import (
	"fmt"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
)

// BenchmarkOperation registers one native benchmark tree from the same
// descriptors/bindings as simdbench. Kernel validation/call boundaries and the
// preallocated per-size buffers retain the legacy benchmark contract.
func BenchmarkOperation(b *testing.B, operationID string) {
	operation, err := kernels.Lookup(operationID, catalog.Float32)
	if err != nil {
		b.Fatal(err)
	}
	if b.Name() != operation.Descriptor.BenchmarkName {
		b.Fatalf("native benchmark %q differs from catalogue alias %q", b.Name(), operation.Descriptor.BenchmarkName)
	}
	var vector simd.Float32s
	for _, size := range BenchmarkSizes(vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			a, inputB, out := MakeBenchmarkSlices(size)
			bytes, err := operation.Descriptor.LogicalBytes(size)
			if err != nil {
				b.Fatal(err)
			}
			for _, implementation := range operation.Implementations {
				b.Run(implementation.Descriptor.Label, func(b *testing.B) {
					Measure(b, size, bytes, out, func() error { return implementation.Run(a, inputB, out) })
				})
			}
		})
	}
}
