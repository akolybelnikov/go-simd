package computation_test

import (
	"fmt"
	"testing"

	"simd"

	"github.com/akolybelnikov/go-simd/packages/benches"
	"github.com/akolybelnikov/go-simd/packages/benches/computation"
)

func TestComputeHeavyScalarAndSIMDAgree(t *testing.T) {
	var vector simd.Float32s
	width := vector.Len()
	for _, size := range []int{0, 1, width - 1, width, width + 1, 2*width + 3} {
		t.Run(fmt.Sprintf("n=%d", size), func(t *testing.T) {
			a, b, scalarOut := benches.MakeBenchmarkSlices(size)
			simdOut := make([]float32, size)

			if err := computation.ComputeHeavyScalar(a, b, scalarOut); err != nil {
				t.Fatalf("ComputeHeavyScalar() error = %v", err)
			}
			if err := computation.ComputeHeavySIMD(a, b, simdOut); err != nil {
				t.Fatalf("ComputeHeavySIMD() error = %v", err)
			}
			for i := range scalarOut {
				if scalarOut[i] != simdOut[i] {
					t.Errorf("SIMD result at index %d = %v, scalar = %v", i, simdOut[i], scalarOut[i])
				}
			}
		})
	}
}

func TestComputeHeavyRejectsMismatchedLengths(t *testing.T) {
	if err := computation.ComputeHeavyScalar([]float32{1}, []float32{}, []float32{0}); err != computation.BadInputError {
		t.Errorf("ComputeHeavyScalar() error = %v, want %v", err, computation.BadInputError)
	}
	if err := computation.ComputeHeavySIMD([]float32{1}, []float32{}, []float32{0}); err != computation.BadInputError {
		t.Errorf("ComputeHeavySIMD() error = %v, want %v", err, computation.BadInputError)
	}
}
