package engine

import (
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

func (Testing) Measure(p Prepared, benchtime string) (Sample, error) {
	if err := (Config{Count: 1, Benchtime: benchtime}).Validate(); err != nil {
		return Sample{}, err
	}
	if p.Run == nil || len(p.Output) == 0 || p.Bytes < 0 {
		return Sample{}, fmt.Errorf("invalid prepared benchmark")
	}
	benchmarkMu.Lock()
	defer benchmarkMu.Unlock()
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
		sink = p.Output[len(p.Output)-1]
	})
	if runErr != nil {
		return Sample{}, runErr
	}
	return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
}
