package main

import "testing"

var quickUint64Sink []uint64

func measureQuickUint64(call func() error, out []uint64, bytes int64) (testing.BenchmarkResult, error) {
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
		quickUint64Sink = out
	})
	return stats, runErr
}
