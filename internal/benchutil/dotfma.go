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
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/input"
	"github.com/akolybelnikov/go-simd/internal/result"
)

// Isolate fused-dot reporting from frozen reduction/pattern callback boundaries.
func benchmarkDotFMA(b *testing.B, op kernels.Operation) {
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
		b.Fatal("native fused-dot catalogue mismatch")
	}
	var vector simd.Float32s
	for _, n := range BenchmarkSizes(vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			x, y, _ := input.MakeForOperation(op.Descriptor, n, options)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, impl := range op.Implementations {
				b.Run(impl.Descriptor.Label, func(b *testing.B) {
					call := impl.BindCheckedScalar(x, y)
					identity, err := result.NewDotFMAIdentity(impl.Descriptor.ID, fmacap.Path(), vector.Len(), n)
					if err != nil {
						b.Fatal(err)
					}
					code, err := result.DotFMAPathCode(identity.PathID)
					if err != nil {
						b.Fatal(err)
					}
					b.ReportAllocs()
					b.SetBytes(bytes)
					b.ResetTimer()
					var value float32
					for i := 0; i < b.N; i++ {
						value, err = call()
						if err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					benchmarkSink = value
					flag := float64(0)
					if fmacap.Path() == fma32.VectorPath {
						flag = 1
					}
					b.ReportMetric(flag, "fma-vector")
					b.ReportMetric(float64(code), "fma-path")
					b.ReportMetric(float64(vector.Len()), "simd-lanes")
					b.ReportMetric(float64(math.Float32bits(value)), "result-bits")
				})
			}
		})
	}
}
