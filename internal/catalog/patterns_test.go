package catalog

import "testing"

func TestPatternDescriptors(t *testing.T) {
	if len(All()) != 7 || len(Groups()) != 7 {
		t.Fatal("operation inventory drift")
	}
	for _, id := range []string{Scale, Affine, ThresholdZero, SelectGT, RangeZero} {
		op, ok := Lookup(id, Float32)
		if !ok {
			t.Fatal(id)
		}
		wantInputs, wantBytes := 1, int64(8*1024)
		if id == SelectGT {
			wantInputs, wantBytes = 2, 12*1024
		}
		bytes, err := op.LogicalBytes(1024)
		if err != nil || bytes != wantBytes || len(op.Inputs) != wantInputs || len(op.Variants) != 2 || !op.Parameterized() || op.NativePath() != "./internal/patterns" {
			t.Fatalf("invalid descriptor: %+v", op)
		}
		for _, p := range op.Parameters {
			if p.TypeID != Float32 || p.Encoding != "ieee754-bits" {
				t.Fatal(p)
			}
		}
	}
}
