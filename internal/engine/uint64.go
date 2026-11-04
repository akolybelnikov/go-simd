package engine

import (
	"context"
	"flag"
	"fmt"
	"testing"
)

var retainedUint64Sink []uint64

// Keep the uint64 callback separate from existing float32 timing contracts.
func measureUint64Context(ctx context.Context, p Prepared, benchtime string) (Sample, error) {
	if p.Run == nil || len(p.Uint64Output) == 0 || p.Float64Output != nil || p.Output != nil || p.ByteOutput != nil || p.Reduce != nil || p.ReduceChecked != nil || p.Scalar != nil || p.Bytes < 0 || !p.RetainOutputAfterTiming {
		return Sample{}, fmt.Errorf("invalid prepared uint64 benchmark")
	}
	benchmarkMu.Lock()
	defer benchmarkMu.Unlock()
	if err := ctx.Err(); err != nil {
		return Sample{}, err
	}
	testing.Init()
	timing := flag.Lookup("test.benchtime")
	previous := timing.Value.String()
	if err := timing.Value.Set(benchtime); err != nil {
		return Sample{}, err
	}
	defer timing.Value.Set(previous)
	var runErr error
	stats := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(p.Bytes)
		b.ResetTimer()
		if runErr != nil {
			b.StopTimer()
			return
		}
		for i := 0; i < b.N; i++ {
			if runErr = p.Run(); runErr != nil {
				b.StopTimer()
				return
			}
		}
		b.StopTimer()
		retainedUint64Sink = p.Uint64Output
	})
	if runErr != nil {
		return Sample{}, runErr
	}
	return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
}
