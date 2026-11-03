package benchutil

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/patterns"
)

// Keep FMA-only metric reporting outside the frozen pattern callbacks.
func benchmarkFMA(b *testing.B, op kernels.Operation) {
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
		b.Fatal("native FMA catalogue name mismatch")
	}
	var vector simd.Float32s
	for _, n := range BenchmarkSizes(vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			x, _, out := input.MakeForOperation(op.Descriptor, n, options)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, impl := range op.Implementations {
				b.Run(impl.Descriptor.Label, func(b *testing.B) {
					call := impl.BindConfigured(x, nil, out, options)
					Measure(b, n, bytes, out, call)
					b.StopTimer()
					flag := float64(0)
					if patterns.FMAPath() == fma32.VectorPath {
						flag = 1
					}
					path := float64(0)
					if impl.Descriptor.UsesSIMD {
						path = flag
					}
					b.ReportMetric(flag, "fma-vector")
					b.ReportMetric(path, "fma-path")
					b.ReportMetric(float64(vector.Len()), "simd-lanes")
					b.ReportMetric(float64(math.Float32bits(out[len(out)-1])), "last-result-bits")
				})
			}
		})
	}
}
