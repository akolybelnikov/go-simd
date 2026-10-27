package kernels

import (
	"errors"
	"fmt"
	"reflect"
	"simd"
	"slices"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/addition"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/computation"
)

func TestExactLegacyKernelBindings(t *testing.T) {
	want := map[bindingKey]Float32Kernel{
		{catalog.Add, catalog.Scalar}:                    addition.AddScalar,
		{catalog.Add, catalog.SIMDPartial}:               addition.AddSIMD,
		{catalog.Add, catalog.SIMDScalarTail}:            addition.AddSIMDVectorTail,
		{catalog.Add, catalog.SIMDParallelScalarTail}:    addition.AddSIMDParallel,
		{catalog.ComputeExpression, catalog.Scalar}:      computation.ComputeHeavyScalar,
		{catalog.ComputeExpression, catalog.SIMDPartial}: computation.ComputeHeavySIMD,
	}
	registered := 0
	for _, descriptor := range catalog.All() {
		if descriptor.Parameterized() {
			continue
		}
		op, err := Lookup(descriptor.ID, descriptor.TypeID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(op.Descriptor, descriptor) || len(op.Implementations) != len(descriptor.Variants) {
			t.Fatal("metadata and runtime registry drift")
		}
		for i, implementation := range op.Implementations {
			key := bindingKey{descriptor.ID, implementation.Descriptor.ID}
			if implementation.Descriptor != descriptor.Variants[i] || want[key] == nil || reflect.ValueOf(implementation.Run).Pointer() != reflect.ValueOf(want[key]).Pointer() {
				t.Fatalf("incorrect typed kernel binding: %+v", key)
			}
			registered++
		}
	}
	if registered != len(want) || len(bindings) != len(want) {
		t.Fatal("missing or orphaned executable bindings")
	}
	for _, pair := range [][2]string{{catalog.Add, "float64"}, {"sum-f32", catalog.Float32}} {
		if _, err := Lookup(pair[0], pair[1]); err == nil {
			t.Fatalf("unavailable pair accepted: %v", pair)
		}
	}
}

func TestBoundAndNativeKernelsPreserveContracts(t *testing.T) {
	var vector simd.Float32s
	w := vector.Len()
	for _, descriptor := range catalog.All() {
		if descriptor.Parameterized() {
			continue
		}
		op, err := Lookup(descriptor.ID, descriptor.TypeID)
		if err != nil {
			t.Fatal(err)
		}
		for _, implementation := range op.Implementations {
			for _, mode := range []string{"native", "bound"} {
				for _, size := range []int{0, 1, w - 1, w, w + 1, 2*w - 1, 2 * w, 2*w + 1, 1025} {
					for _, alias := range []string{"none", "a", "b"} {
						t.Run(fmt.Sprintf("%s/%s/%s/n=%d/%s", descriptor.ID, implementation.Descriptor.Label, mode, size, alias), func(t *testing.T) {
							a, b, out, want := make([]float32, size), make([]float32, size), make([]float32, size), make([]float32, size)
							for i := range a {
								x, y := float32(i%11-5), float32(i%7-3)
								a[i], b[i] = x, y
								want[i] = x + y
								if descriptor.ID == catalog.ComputeExpression {
									want[i] = x*x + y*y + x*y
								}
							}
							if alias == "a" {
								out = a
							}
							if alias == "b" {
								out = b
							}
							var err error
							if mode == "native" {
								err = implementation.Run(a, b, out)
							} else {
								err = implementation.Bind(a, b, out)()
							}
							if err != nil || !slices.Equal(out, want) {
								t.Fatalf("result=%v error=%v, want %v", out, err, want)
							}
						})
					}
				}
				out := []float32{123}
				var mismatch error
				if mode == "native" {
					mismatch = implementation.Run([]float32{1, 2}, []float32{3, 4}, out)
				} else {
					mismatch = implementation.Bind([]float32{1, 2}, []float32{3, 4}, out)()
				}
				wantErr := addition.ErrLengthMismatch
				if descriptor.ID == catalog.ComputeExpression {
					wantErr = computation.ErrLengthMismatch
				}
				if !errors.Is(mismatch, wantErr) || out[0] != 123 {
					t.Fatal("length-validation or unchanged-output contract lost")
				}
			}
		}
	}
}
