package benchutil

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
)

var xorBitsOutputSink []float32

func benchmarkXorBits(b *testing.B, op kernels.Operation) {
	if b.Name() != op.Descriptor.BenchmarkName {
		b.Fatal("native XOR bits catalogue mismatch")
	}
	var vector simd.Float32s
	for _, n := range OperationSizes(op.Descriptor, vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			x, y, out := input.MakeForOperation(op.Descriptor, n, input.DefaultOptions())
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, impl := range op.Implementations {
				b.Run(impl.Descriptor.Label, func(b *testing.B) {
					call := impl.Bind(x, y, out)
					b.ReportAllocs()
					b.SetBytes(bytes)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if err := call(); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					xorBitsOutputSink = out
					digest, _ := hex.DecodeString(input.FingerprintFloat32(out, nil))
					for i := 0; i < 8; i++ {
						b.ReportMetric(float64(binary.BigEndian.Uint32(digest[i*4:])), fmt.Sprintf("output-sha256-%d", i))
					}
					b.ReportMetric(float64(vector.Len()), "simd-lanes")
				})
			}
		})
	}
}
