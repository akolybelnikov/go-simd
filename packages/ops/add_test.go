package ops_test

import (
	"testing"

	"github.com/akolybelnikov/go-simd/packages/ops"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		a    []float32
		b    []float32
		want []float32
	}{
		{
			name: "Test Add exactly one simd vector",
			a:    []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0},
			b:    []float32{5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0},
			want: []float32{6.0, 8.0, 10.0, 12.0, 14.0, 16.0, 18.0, 20.0},
		},
		{
			name: "Test Add Tail",
			a:    []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0},
			b:    []float32{5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0},
			want: []float32{6.0, 8.0, 10.0, 12.0, 14.0, 16.0, 18.0, 20.0, 22.0, 24.0, 26.0},
		},
		{
			name: "Test Add Empty",
			a:    []float32{},
			b:    []float32{},
			want: []float32{},
		},
		{
			name: "Fewer values",
			a:    []float32{1.0, 2.0, 3.0},
			b:    []float32{4.0, 5.0},
			want: []float32{5.0, 7.0, 3.0},
		},
		{
			name: "Negative and zero values",
			a:    []float32{-1.0, 0.0, 3.0},
			b:    []float32{4.0, -5.0, 2.0},
			want: []float32{3.0, -5.0, 5.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]float32, len(tt.want))
			ops.Add(tt.a, tt.b, out)
			for i := range out {
				if out[i] != tt.want[i] {
					t.Errorf("Add() = %v, want %v", out, tt.want)
					break
				}
			}
		})
	}
}
