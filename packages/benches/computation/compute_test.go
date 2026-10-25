package computation_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/packages/computation"
)

func TestComputeHeavyRejectsMismatchedLengths(t *testing.T) {
	if err := computation.ComputeHeavyScalar([]float32{1}, []float32{}, []float32{0}); err != computation.BadInputError {
		t.Errorf("ComputeHeavyScalar() error = %v, want %v", err, computation.BadInputError)
	}
	if err := computation.ComputeHeavySIMD([]float32{1}, []float32{}, []float32{0}); err != computation.BadInputError {
		t.Errorf("ComputeHeavySIMD() error = %v, want %v", err, computation.BadInputError)
	}
}
