package engine

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"
)

// Testing serializes the documented testing flag API and restores the caller's
// setting. Do not concurrently run unrelated testing.Benchmark users in-process.
type Testing struct{}

var benchmarkMu sync.Mutex
var sink float32
var byteSink []uint8

func (Testing) Measure(p Prepared, benchtime string) (Sample, error) {
	return (Testing{}).MeasureContext(context.Background(), p, benchtime)
}

// MeasureContext observes cancellation before and after acquiring the shared
// testing configuration lock. It intentionally does not add context checks to
// the timed benchmark loop; ExecuteContext observes cancellation after a
// completed sample, while an isolated worker can be terminated for a hard stop.
func (Testing) MeasureContext(ctx context.Context, p Prepared, benchtime string) (Sample, error) {
	if ctx == nil {
		return Sample{}, fmt.Errorf("nil measurement context")
	}
	if err := ctx.Err(); err != nil {
		return Sample{}, err
	}
	if err := (Config{Count: 1, Benchtime: benchtime}).Validate(); err != nil {
		return Sample{}, err
	}
	scalar := p.Reduce != nil || p.ReduceChecked != nil
	bytes := p.ByteOutput != nil
	if p.RetainOutputAfterTiming && (bytes || scalar || p.Scalar != nil) {
		return Sample{}, fmt.Errorf("invalid retained array benchmark")
	}
	if p.Bytes < 0 || (bytes && (scalar || p.Scalar != nil || p.Output != nil || p.Run == nil || len(p.ByteOutput) == 0)) || (!bytes && !scalar && (p.Run == nil || len(p.Output) == 0)) || (scalar && (p.Scalar == nil || p.Run != nil || p.Output != nil)) || (p.Reduce != nil && p.ReduceChecked != nil) {
		return Sample{}, fmt.Errorf("invalid prepared benchmark")
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
	if p.RetainOutputAfterTiming {
		return measureRetainedArray(p)
	}
	if bytes {
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
			byteSink = p.ByteOutput
		})
		if runErr != nil {
			return Sample{}, runErr
		}
		return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
	}
	if p.ReduceChecked != nil {
		var runErr error
		stats := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(p.Bytes)
			b.ResetTimer()
			if runErr != nil {
				b.StopTimer()
				return
			}
			var value float32
			for i := 0; i < b.N; i++ {
				value, runErr = p.ReduceChecked()
				if runErr != nil {
					b.StopTimer()
					return
				}
			}
			b.StopTimer()
			*p.Scalar = value
			sink = value
		})
		if runErr != nil {
			return Sample{}, runErr
		}
		return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
	}
	if p.Reduce != nil {
		stats := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(p.Bytes)
			b.ResetTimer()
			var value float32
			for i := 0; i < b.N; i++ {
				value = p.Reduce()
			}
			b.StopTimer()
			*p.Scalar = value
			sink = value
		})
		return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
	}
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
		sink = p.Output[len(p.Output)-1]
	})
	if runErr != nil {
		return Sample{}, runErr
	}
	return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
}
