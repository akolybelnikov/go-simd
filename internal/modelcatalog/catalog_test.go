package modelcatalog

import "testing"

func TestAssessmentDoesNotPromoteNarrowOrUnrunEvidence(t *testing.T) {
	if len(BrowserRecommendations()) != 0 {
		t.Fatal("current evaluations do not support browser recommendations")
	}
	for _, tag := range []string{"qwen3:4b", "qwen3.5:4b", "granite4.2:3b", "qwen3.5:9b"} {
		a := Browser(tag)
		if a.Recommended || a.Evaluated != "2026-11-22" || a.EvidenceURL == "" {
			t.Fatalf("invalid browser assessment for %s: %+v", tag, a)
		}
	}
	for _, tag := range []string{"inkling-2-mini:latest", "an-unknown-model", "qwen3.5:9b-another-quantization"} {
		a := Browser(tag)
		if a.Recommended || a.Evaluated != "" {
			t.Fatalf("must not inherit tested status: %s: %+v", tag, a)
		}
	}
}
