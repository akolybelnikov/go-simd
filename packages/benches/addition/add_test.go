package addition_test

import (
	"errors"
	"fmt"
	"testing"

	"simd"

	"github.com/akolybelnikov/go-simd/packages/addition"
)

var tests = []struct {
	name string
	a    []float32
	b    []float32
	want []float32
	wantErr error
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
		name:    "Fewer values",
		a:       []float32{1.0, 2.0, 3.0},
		b:       []float32{4.0, 5.0},
		want:    []float32{5.0, 7.0, 3.0},
		wantErr: addition.BadInputError,
	},
	{
		name: "Negative and zero values",
		a:    []float32{-1.0, 0.0, 3.0},
		b:    []float32{4.0, -5.0, 2.0},
		want: []float32{3.0, -5.0, 5.0},
	},
}

func TestAddSIMD(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]float32, len(tt.want))
			if err := addition.AddSIMD(tt.a, tt.b, out); err != nil {
				if tt.wantErr == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("AddSIMD() error = %v", err)
				}
				return
			}
			for i := range out {
				if out[i] != tt.want[i] {
					t.Errorf("AddSIMD() = %v, want %v", out, tt.want)
					break
				}
			}
		})
	}
}

func TestAddScalar(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make([]float32, len(tt.want))
			if err := addition.AddScalar(tt.a, tt.b, out); err != nil {
				if tt.wantErr == nil || !errors.Is(err, tt.wantErr) {
					t.Errorf("AddScalar() error = %v", err)
				}
				return
			}
			for i := range out {
				if out[i] != tt.want[i] {
					t.Errorf("AddScalar() = %v, want %v", out, tt.want)
					break
				}
			}
		})
	}
}

func TestAddImplementationsAcrossVectorBoundaries(t *testing.T) {
	implementations := []struct {
		name string
		add  func([]float32, []float32, []float32) error
	}{
		{name: "SIMDPartial", add: addition.AddSIMD},
		{name: "SIMDTail", add: addition.AddSIMDVectorTail},
		{name: "SIMDParallel", add: addition.AddSIMDParallel},
	}

	var vector simd.Float32s
	width := vector.Len()
	sizes := []int{0, 1, width - 1, width, width + 1, 2*width + 3}

	for _, implementation := range implementations {
		t.Run(implementation.name, func(t *testing.T) {
			for _, size := range sizes {
				t.Run(fmt.Sprintf("n=%d", size), func(t *testing.T) {
					a := make([]float32, size)
					b := make([]float32, size)
					out := make([]float32, size)
					for i := range a {
						a[i] = float32(i - 3)
						b[i] = float32(2*i + 1)
					}

					if err := implementation.add(a, b, out); err != nil {
						t.Fatalf("Add() error = %v", err)
					}
					for i := range out {
						want := a[i] + b[i]
						if out[i] != want {
							t.Errorf("out[%d] = %v, want %v", i, out[i], want)
						}
					}
				})
			}
		})
	}
}

func TestAddSIMDVariantsRejectMismatchedLengths(t *testing.T) {
	implementations := []struct {
		name string
		add  func([]float32, []float32, []float32) error
	}{
		{name: "SIMDPartial", add: addition.AddSIMD},
		{name: "SIMDTail", add: addition.AddSIMDVectorTail},
		{name: "SIMDParallel", add: addition.AddSIMDParallel},
	}

	for _, implementation := range implementations {
		t.Run(implementation.name, func(t *testing.T) {
			err := implementation.add([]float32{1, 2}, []float32{3}, make([]float32, 2))
			if !errors.Is(err, addition.BadInputError) {
				t.Fatalf("Add() error = %v, want %v", err, addition.BadInputError)
			}
		})
	}
}
