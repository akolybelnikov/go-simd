package engine

import "testing"

var retainedArraySink []float32

// Caller owns the shared testing lock and configured/restored benchtime.
// Keep post-timing array retention separate from older callback boundaries.
func measureRetainedArray(p Prepared) (Sample, error) {
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
		retainedArraySink = p.Output
	})
	if runErr != nil {
		return Sample{}, runErr
	}
	return Sample{stats.N, stats.T, stats.MemAllocs, stats.MemBytes}, nil
}
