package addition

import (
	"errors"
	"runtime"
	"simd"
	"sync"
)

// ErrLengthMismatch indicates unequal input and output lengths.
var ErrLengthMismatch = errors.New("input and output slices must have equal lengths")

// AddSIMD adds inputs using partial SIMD loads and stores.
func AddSIMD(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := 0; i < len(a); {
		x, n := simd.LoadFloat32sPart(a[i:])
		y, _ := simd.LoadFloat32sPart(b[i:])
		x.Add(y).StorePart(out[i:])
		i += n
	}

	return nil
}

// AddSIMDVectorTail adds full SIMD vectors, then handles any remaining
// elements with scalar additions.
func AddSIMDVectorTail(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	addSIMDVectorTail(a, b, out)
	return nil
}

// AddSIMDParallel divides the input into contiguous chunks and adds each
// chunk using SIMD vectors with a scalar tail. It starts at most GOMAXPROCS
// workers; small inputs may be slower because of goroutine overhead.
func AddSIMDParallel(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	if len(a) == 0 {
		return nil
	}

	workers := min(runtime.GOMAXPROCS(0), len(a))
	chunkSize := len(a) / workers
	extra := len(a) % workers

	var wg sync.WaitGroup
	wg.Add(workers)
	for worker := 0; worker < workers; worker++ {
		start := worker*chunkSize + min(worker, extra)
		end := start + chunkSize
		if worker < extra {
			end++
		}
		go func() {
			defer wg.Done()
			addSIMDVectorTail(a[start:end], b[start:end], out[start:end])
		}()
	}
	wg.Wait()
	return nil
}

func addSIMDVectorTail(a, b, out []float32) {
	var vector simd.Float32s
	width := vector.Len()
	i := 0
	for ; i+width <= len(a); i += width {
		x := simd.LoadFloat32s(a[i:])
		y := simd.LoadFloat32s(b[i:])
		x.Add(y).Store(out[i:])
	}
	for ; i < len(a); i++ {
		out[i] = a[i] + b[i]
	}
}

// AddScalar adds inputs using a scalar loop.
func AddScalar(a, b, out []float32) error {
	if len(a) != len(b) || len(a) != len(out) {
		return ErrLengthMismatch
	}
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return nil
}
