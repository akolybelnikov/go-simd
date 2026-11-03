package benchutil

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
)

func benchmarkReduction(b *testing.B, op kernels.Operation) {
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
			x, y, _ := input.MakeForOperation(op.Descriptor, n, options)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, implementation := range op.Implementations {
				b.Run(implementation.Descriptor.Label, func(b *testing.B) {
					if implementation.BindCheckedScalar != nil {
						call := implementation.BindCheckedScalar(x, y)
						b.ReportAllocs()
						b.SetBytes(bytes)
						b.ResetTimer()
						var value float32
						for i := 0; i < b.N; i++ {
							var err error
							value, err = call()
							if err != nil {
								b.Fatal(err)
							}
						}
						b.StopTimer()
						benchmarkSink = value
						b.ReportMetric(float64(vector.Len()), "simd-lanes")
						b.ReportMetric(float64(math.Float32bits(value)), "result-bits")
						return
					}
					call := implementation.BindScalar(x)
					b.ReportAllocs()
					b.SetBytes(bytes)
					b.ResetTimer()
					var value float32
					for i := 0; i < b.N; i++ {
						value = call()
					}
					b.StopTimer()
					benchmarkSink = value
					b.ReportMetric(float64(vector.Len()), "simd-lanes")
					b.ReportMetric(float64(math.Float32bits(value)), "result-bits")
				})
			}
		})
	}
}
