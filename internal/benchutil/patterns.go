package benchutil

import (
	"encoding/json"
	"fmt"
	"os"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
)

func benchmarkPattern(b *testing.B, op kernels.Operation) {
	options := input.DefaultOptions()
	if text, ok := os.LookupEnv(input.NativeOptionsEnv); ok {
		if err := json.Unmarshal([]byte(text), &options); err != nil {
			b.Fatal(err)
		}
	}
	if err := options.ValidateFor(op.Descriptor); err != nil {
		b.Fatal(err)
	}
	if b.Name() != op.Descriptor.BenchmarkName {
		b.Fatal("native catalogue name mismatch")
	}
	var vector simd.Float32s
	for _, n := range BenchmarkSizes(vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			a, other, out := input.MakeForOperation(op.Descriptor, n, options)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, implementation := range op.Implementations {
				b.Run(implementation.Descriptor.Label, func(b *testing.B) {
					call := implementation.BindConfigured(a, other, out, options)
					Measure(b, n, bytes, out, call)
				})
			}
		})
	}
}
