package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/platform"
)

func TestCompactCatalogue(t *testing.T) {
	var out bytes.Buffer
	if err := PrintCatalog(&out, "all", platform.Info{}); err != nil {
		t.Fatal(err)
	}
	for _, op := range catalog.All() {
		if !strings.Contains(out.String(), op.Group) || !strings.Contains(out.String(), "["+op.ID+"/float32]") {
			t.Fatal("missing compact operation")
		}
	}
	if strings.Contains(out.String(), "scalar reference:") || !strings.Contains(out.String(), "Use -list -group") || strings.Count(out.String(), "\n") > 25 {
		t.Fatal("default catalogue is not compact")
	}
}
