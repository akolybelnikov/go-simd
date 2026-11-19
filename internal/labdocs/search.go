// Package labdocs searches a curated, compiled set of lab documentation notes.
package labdocs

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

const (
	maxQueryBytes = 256
	maxResults    = 3
)

// Result contains a curated note and its repository source location.
type Result struct {
	Source  string `json:"source"`
	Section string `json:"section"`
	Summary string `json:"summary"`
}

var approved = []struct {
	keywords string
	result   Result
}{
	{"addition add scalar simdpartial simdtail simdparallel implementations strategies kernel", Result{
		Source: "docs/wiki/04-Operation-guide.md", Section: "Complete kernel reference",
		Summary: "Float32 addition has scalar, partial-vector, full-vector with scalar tail, and parallel SIMD implementations. Parallel chunks each use full vectors and a scalar tail.",
	}},
	{"measurements results latency bandwidth logical bytes timing ns op allocation", Result{
		Source: "docs/wiki/06-Understanding-results.md", Section: "Read each number correctly",
		Summary: "ns/op is average time per complete kernel call over the selected slice. Effective MB/s is based on logical work accounting. It does not measure DRAM bandwidth.",
	}},
	{"sizes lanes vector width boundaries tail remainder supported emulation", Result{
		Source: "docs/wiki/06-Understanding-results.md", Section: "Vector boundaries and tails",
		Summary: "Use the recorded lane count to interpret vector-boundary sizes. A boundary observed at one width does not transfer to another width. Emulation and lowering depend on the runtime environment.",
	}},
	{"parallel serial workers vectorization speedup confounding gomaxprocs", Result{
		Source: "docs/wiki/06-Understanding-results.md", Section: "Serial versus parallel comparisons",
		Summary: "SIMDParallel adds concurrency and synchronization. Its comparison with the serial scalar baseline does not isolate vectorization. GOMAXPROCS affects the experiment.",
	}},
}

// Search returns up to three notes ranked by the number of matching distinct
// keywords. The query cannot select a file or a network location.
func Search(query string) ([]Result, error) {
	if len(query) > maxQueryBytes {
		return nil, fmt.Errorf("documentation query exceeds %d bytes", maxQueryBytes)
	}
	terms := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if len(terms) == 0 {
		return nil, fmt.Errorf("documentation query needs at least one word")
	}
	slices.Sort(terms)
	terms = slices.Compact(terms)
	type match struct {
		result Result
		score  int
	}
	var matches []match
	for _, entry := range approved {
		keywords := strings.Fields(entry.keywords)
		score := 0
		for _, term := range terms {
			if slices.Contains(keywords, term) {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, match{entry.result, score})
		}
	}
	slices.SortStableFunc(matches, func(a, b match) int { return b.score - a.score })
	results := make([]Result, 0, min(len(matches), maxResults))
	for _, item := range matches[:min(len(matches), maxResults)] {
		results = append(results, item.result)
	}
	return results, nil
}
