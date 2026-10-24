package ops_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/packages/ops"
)

func TestComputeHeavyRejectsMismatchedLengths(t *testing.T) {
	if err := ops.ComputeHeavyScalar([]float32{1}, []float32{}, []float32{0}); err != ops.BadInputError {
		t.Errorf("ComputeHeavyScalar() error = %v, want %v", err, ops.BadInputError)
	}
	if err := ops.ComputeHeavySIMD([]float32{1}, []float32{}, []float32{0}); err != ops.BadInputError {
		t.Errorf("ComputeHeavySIMD() error = %v, want %v", err, ops.BadInputError)
	}
}
