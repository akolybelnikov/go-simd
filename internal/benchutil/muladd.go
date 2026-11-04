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
	"github.com/akolybelnikov/go-simd/internal/fma32"
	"github.com/akolybelnikov/go-simd/internal/fmacap"
	"github.com/akolybelnikov/go-simd/internal/input"
)

var muladdOutputSink []float32

func benchmarkMulAdd(b *testing.B, op kernels.Operation) {
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
		b.Fatal("native muladd catalogue mismatch")
	}
	var vector simd.Float32s
	for _, n := range OperationSizes(op.Descriptor, vector.Len()) {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			x, y, z, out := input.MakeMulAdd(n, o)
			bytes, err := op.Descriptor.LogicalBytes(n)
			if err != nil {
				b.Fatal(err)
			}
			for _, impl := range op.Implementations {
				b.Run(impl.Descriptor.Label, func(b *testing.B) {
					call := impl.BindTernary(x, y, z, out)
					b.ReportAllocs()
					b.SetBytes(bytes)
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if err := call(); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					muladdOutputSink = out
					digest, _ := hex.DecodeString(input.FingerprintFloat32(out, nil))
					for i := 0; i < 8; i++ {
						b.ReportMetric(float64(binary.BigEndian.Uint32(digest[i*4:])), fmt.Sprintf("output-sha256-%d", i))
					}
					b.ReportMetric(float64(vector.Len()), "simd-lanes")
					gate, path := float64(0), float64(0)
					if fmacap.Path() == fma32.VectorPath {
						gate = 1
						if impl.Descriptor.UsesSIMD {
							path = 1
						}
					}
					b.ReportMetric(gate, "fma-vector")
					b.ReportMetric(path, "fma-path")
				})
			}
		})
	}
}
