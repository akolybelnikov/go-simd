package labeval

import (
	"slices"
	"testing"
)

func TestCorpusIsVersionedUniqueAndCoversAllWorkflows(t *testing.T) {
	corpus, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) < 8 {
		t.Fatalf("evaluation coverage unexpectedly small: %d", len(corpus.Cases))
	}
}

func TestEvaluateRejectsAuthorityAndClaimViolations(t *testing.T) {
	testCase := Case{
		RequiredTools: []string{"inspect_evidence"}, ForbiddenTools: []string{"shell"},
		RequiredCaveats: []string{"low_sample_count"}, ForbiddenClaims: []string{"statistical_significance"},
		RequireNumericFactRefs: true,
	}
	trace := Trace{
		ToolCalls: []string{"shell"}, ClaimClasses: []string{"statistical_significance"},
		NumericClaims: []NumericClaim{{Text: "12% faster"}},
	}
	failures := Evaluate(testCase, trace)
	for _, code := range []string{"missing_tool", "forbidden_tool", "missing_caveat", "forbidden_claim", "untraceable_numeric_claim"} {
		if !slices.ContainsFunc(failures, func(f Failure) bool { return f.Code == code }) {
			t.Errorf("missing failure %q in %+v", code, failures)
		}
	}
}

func TestEvaluateAcceptsGroundedTrace(t *testing.T) {
	testCase := Case{RequiredTools: []string{"inspect_evidence"}, RequiredCaveats: []string{"low_sample_count"}, RequireNumericFactRefs: true}
	trace := Trace{ToolCalls: []string{"inspect_evidence"}, CaveatCodes: []string{"low_sample_count"}, NumericClaims: []NumericClaim{{Text: "median 10 ns/op", FactID: "series-1"}}}
	if failures := Evaluate(testCase, trace); len(failures) != 0 {
		t.Fatalf("grounded trace failures: %+v", failures)
	}
}
