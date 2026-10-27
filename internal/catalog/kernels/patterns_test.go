package kernels

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
)

func TestConfiguredBindingInventory(t *testing.T) {
	count := 0
	for _, op := range catalog.All() {
		if !op.Parameterized() {
			continue
		}
		runtime, err := Lookup(op.ID, op.TypeID)
		if err != nil {
			t.Fatal(err)
		}
		for i, impl := range runtime.Implementations {
			if impl.Descriptor != op.Variants[i] || impl.BindConfigured == nil || impl.Run != nil || impl.Bind != nil {
				t.Fatal("configured binding drift")
			}
			count++
		}
	}
	if count != 10 || len(configuredBindings) != 10 {
		t.Fatal("missing/orphan configured binding")
	}
}
