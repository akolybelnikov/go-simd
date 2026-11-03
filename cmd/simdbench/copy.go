package main

import "testing"

var quickCopySink []float32

func measureQuickCopy(call func() error, out []float32, bytes int64) (testing.BenchmarkResult, error) {
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
		quickCopySink = out
	})
	return stats, runErr
}
