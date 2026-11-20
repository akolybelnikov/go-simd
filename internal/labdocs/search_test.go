package labdocs

import (
	"strings"
	"testing"
)

func TestSearchBoundsAndReferences(t *testing.T) {
	results, err := Search("SIMDParallel addition strategies")
	if err != nil || len(results) == 0 || results[0].Source != "docs/wiki/04-Operation-guide.md" || results[0].Section != "Addition methods" {
		t.Fatalf("unexpected search results: %+v, %v", results, err)
	}
	results, err = Search("../../secrets.txt")
	if err != nil || len(results) != 0 {
		t.Fatalf("unexpected matches for unlisted terms: %+v, %v", results, err)
	}
	results, err = Search("addition vector timing parallel")
	if err != nil || len(results) != maxResults {
		t.Fatalf("result cap failed: %+v, %v", results, err)
	}
	if _, err := Search(strings.Repeat("x", maxQueryBytes+1)); err == nil {
		t.Fatal("oversized query accepted")
	}
}
