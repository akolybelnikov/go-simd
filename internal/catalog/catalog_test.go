package catalog

import (
	"math"
	"reflect"
	"testing"
)

func TestFrozenOperationIdentities(t *testing.T) {
	want := [][]string{
		{"add", "addition", "BenchmarkAddition", "float32", "go-float32-add-v1", "scalar", "Scalar", "simd-partial", "SIMDPartial", "simd-scalar-tail", "SIMDTail", "simd-parallel-scalar-tail", "SIMDParallel"},
		{"compute-expression", "computation", "BenchmarkComputation", "float32", "go-float32-quadratic-v1", "scalar", "Scalar", "simd-partial", "SIMD"},
	}
	var got [][]string
	for _, op := range All() {
		row := []string{op.ID, op.Group, op.BenchmarkName, op.TypeID, op.ArithmeticID}
		if op.ReferenceVariantID != Scalar || op.ExecutionMode != KernelOnly || len(op.Inputs) != 2 || len(op.Parameters) != 0 || op.Output.Shape != Array {
			t.Fatalf("unexpected existing-operation contract: %+v", op)
		}
		seen := map[string]bool{}
		for _, variant := range op.Variants {
			if seen[variant.ID] {
				t.Fatalf("duplicate strategy %s/%s", op.ID, variant.ID)
			}
			seen[variant.ID] = true
			row = append(row, variant.ID, variant.Label)
		}
		got = append(got, row)
		byGroup, ok := ByGroup(op.Group)
		if !ok || !reflect.DeepEqual(op, byGroup) {
			t.Fatal("group lookup drift")
		}
		byBenchmark, ok := ByBenchmark(op.BenchmarkName)
		if !ok || !reflect.DeepEqual(op, byBenchmark) {
			t.Fatal("native alias lookup drift")
		}
		if bytes, err := op.LogicalBytes(1024); err != nil || bytes != 12288 {
			t.Fatalf("legacy traffic = %d, %v", bytes, err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identities/aliases changed\ngot: %v\nwant: %v", got, want)
	}
	if _, ok := Lookup(Add, "float64"); ok {
		t.Fatal("unavailable type silently accepted")
	}
	if _, ok := Lookup("broadcast", Float32); ok {
		t.Fatal("planned operation advertised as implemented")
	}
	if _, ok := ByGroup("unknown"); ok {
		t.Fatal("unknown group accepted")
	}
	if _, ok := ByBenchmark("BenchmarkUnknown"); ok {
		t.Fatal("unknown benchmark accepted")
	}
}

func TestDescriptorsAreIndependentCopies(t *testing.T) {
	before := All()
	modified := All()
	modified[0].ID = "changed"
	modified[0].Inputs[0].TypeID = "changed"
	modified[0].Variants[0].ID = "changed"
	modified[0].Output.ElementBytes = 100
	lookup, ok := Lookup(Add, Float32)
	if !ok {
		t.Fatal("registered operation missing")
	}
	lookup.Inputs[0].Name = "changed"
	lookup.Variants[0].Label = "changed"
	if !reflect.DeepEqual(All(), before) {
		t.Fatal("caller mutated shared catalogue")
	}
}

func TestLogicalTrafficShapes(t *testing.T) {
	a := Value{"a", Float32, Array, 4}
	out := Value{"out", Float32, Array, 4}
	for _, tt := range []struct {
		name   string
		inputs []Value
		output Value
		want   int64
	}{
		{"unary", []Value{a}, out, 80},
		{"binary", []Value{a, a}, out, 120},
		{"ternary", []Value{a, a, a}, out, 160},
		{"sum", []Value{a}, Value{"sum", Float32, ScalarValue, 4}, 44},
		{"dot", []Value{a, a}, Value{"dot", Float32, ScalarValue, 4}, 84},
		{"logical predicate", []Value{a}, Value{"mask", "uint8", Array, 1}, 50},
	} {
		t.Run(tt.name, func(t *testing.T) {
			op := Operation{Inputs: tt.inputs, Output: tt.output, Parameters: []Parameter{{"alpha", Float32, "ieee754-bits"}}}
			if bytes, err := op.LogicalBytes(10); err != nil || bytes != tt.want {
				t.Fatalf("traffic = %d, %v; want %d", bytes, err, tt.want)
			}
		})
	}
	op := Operation{Inputs: []Value{a}, Output: out}
	if bytes, err := op.LogicalBytes(0); err != nil || bytes != 0 {
		t.Fatal("zero work incorrectly counted")
	}
	if _, err := op.LogicalBytes(-1); err == nil {
		t.Fatal("negative length accepted")
	}
	op.Inputs[0].ElementBytes = math.MaxInt64
	if _, err := op.LogicalBytes(2); err == nil {
		t.Fatal("overflow accepted")
	}
	op.Inputs[0].ElementBytes = 0
	if _, err := op.LogicalBytes(2); err == nil {
		t.Fatal("unknown element width accepted")
	}
	op.Inputs[0] = a
	op.Output.Shape = "unknown"
	if _, err := op.LogicalBytes(2); err == nil {
		t.Fatal("unknown shape accepted")
	}
}

func TestWorkerPoliciesAndLegacyOrdering(t *testing.T) {
	for _, tt := range []struct {
		policy         WorkerPolicy
		n, procs, want int
		known          bool
	}{
		{Serial, 0, 0, 1, true}, {Serial, 17, 0, 1, true},
		{PerCallGOMAXPROCS, 0, 0, 0, true}, {PerCallGOMAXPROCS, 1, 8, 1, true},
		{PerCallGOMAXPROCS, 17, 4, 4, true}, {PerCallGOMAXPROCS, 17, 0, 0, false},
		{Serial, -1, 4, 0, false}, {"unknown", 4, 4, 0, false},
	} {
		count, known := (Variant{Workers: tt.policy}).WorkerCount(tt.n, tt.procs)
		if count != tt.want || known != tt.known {
			t.Errorf("%s/%d/%d = %d,%t", tt.policy, tt.n, tt.procs, count, known)
		}
	}
	for rank, label := range []string{"Scalar", "SIMD", "SIMDPartial", "SIMDTail", "SIMDParallel"} {
		if LegacyLabelOrder(label) != rank {
			t.Errorf("legacy order changed for %s", label)
		}
	}
	if LegacyLabelOrder("unknown") <= LegacyLabelOrder("SIMDParallel") {
		t.Fatal("unknown labels must follow known labels")
	}
}
