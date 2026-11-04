package benchutil

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"simd"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog/kernels"
	"github.com/akolybelnikov/go-simd/internal/input"
)

var copy64OutputSink []float64

func benchmarkCopy64(b *testing.B, op kernels.Operation) {
	o := input.DefaultOptions()
	if text, ok := os.LookupEnv(input.NativeOptionsEnv); ok {
		if err := json.Unmarshal([]byte(text), &o); err != nil {
			b.Fatal(err)
		}
	}
	if err := o.ValidateFor(op.Descriptor); err != nil {
		b.Fatal(err)
	}
	if b.Name() != op.Descriptor.BenchmarkName {
		b.Fatal("native copy64 catalogue mismatch")
	}
	var vector simd.Float32s
	for _, n := range OperationSizes(op.Descriptor, vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			x, out := input.MakeCopy64(n)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, impl := range op.Implementations {
				b.Run(impl.Descriptor.Label, func(b *testing.B) {
					call := impl.BindUnaryFloat64(x, out)
					b.ReportAllocs()
					b.SetBytes(bytes)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if err := call(); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					copy64OutputSink = out
					digest, _ := hex.DecodeString(input.FingerprintFloat64(out, nil))
					for i := 0; i < 8; i++ {
						b.ReportMetric(float64(binary.BigEndian.Uint32(digest[i*4:])), fmt.Sprintf("output-sha256-%d", i))
					}
					var actual simd.Float64s
					b.ReportMetric(float64(actual.Len()), "simd-lanes")
				})
			}
		})
	}
}
