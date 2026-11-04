package main

import "testing"

var quickFloat64Sink []float64

func measureQuickFloat64(call func() error, out []float64, bytes int64) (testing.BenchmarkResult, error) {
	var runErr error
	stats := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(bytes)
		b.ResetTimer()
		if runErr != nil {
			b.StopTimer()
			return
		}
		for i := 0; i < b.N; i++ {
			if runErr = call(); runErr != nil {
				b.StopTimer()
				return
			}
		}
		b.StopTimer()
		quickFloat64Sink = out
	})
	return stats, runErr
}
