package labapi

import (
	"encoding/json"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestDescribeCatalogCompletenessAndIsolation(t *testing.T) {
	response, err := DescribeCatalog("all", testRuntime(), "software-fma32/v1")
	if err != nil {
		t.Fatal(err)
	}
	wantOperations, wantStrategies := len(catalog.All()), 0
	for _, op := range catalog.All() {
		wantStrategies += len(op.Variants)
	}
	gotStrategies := 0
	identities := make(map[string]bool)
	for _, op := range response.Operations {
		if len(op.SupportedSizes) == 0 || op.ReferenceStrategy == "" || op.ArithmeticID == "" {
			t.Fatalf("incomplete operation: %+v", op)
		}
		for _, strategy := range op.Strategies {
			identity := op.ID + "/" + op.TypeID + "/" + strategy.ID
			if identities[identity] {
				t.Fatalf("duplicate identity %q", identity)
			}
			identities[identity] = true
			gotStrategies++
		}
	}
	if len(response.Operations) != wantOperations || gotStrategies != wantStrategies {
		t.Fatalf("operations=%d/%d strategies=%d/%d", len(response.Operations), wantOperations, gotStrategies, wantStrategies)
	}
	response.Operations[0].Strategies[0].Label = "mutated"
	second, _ := DescribeCatalog("all", testRuntime(), "software-fma32/v1")
	if second.Operations[0].Strategies[0].Label == "mutated" {
		t.Fatal("catalog response aliases registry data")
	}
}

func TestDescribeCatalogFilterAndJSONContract(t *testing.T) {
	response, err := DescribeCatalog("addition", testRuntime(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Operations) != 1 || response.Operations[0].Group != "addition" {
		t.Fatalf("response=%+v", response)
	}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var header struct {
		Schema string `json:"schema"`
	}
	if json.Unmarshal(data, &header) != nil || header.Schema != CatalogSchema {
		t.Fatalf("catalog JSON missing schema: %s", data)
	}
	if _, err := DescribeCatalog("missing", testRuntime(), ""); errorCode(err) != "unknown_group" {
		t.Fatalf("unknown group error=%v", err)
	}
}
