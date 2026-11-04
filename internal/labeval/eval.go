// Package labeval defines the provider-neutral agent evaluation corpus and
// validates structured traces. It does not use a model as its own judge.
package labeval

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
)

const Schema = "go-simd.agent-eval/v1"

//go:embed corpus.json
var corpusData []byte

type Corpus struct {
	Schema string `json:"schema"`
	Cases  []Case `json:"cases"`
}

type Case struct {
	ID                     string   `json:"id"`
	Workflow               string   `json:"workflow"`
	Prompt                 string   `json:"prompt"`
	RequiredTools          []string `json:"required_tools,omitempty"`
	ForbiddenTools         []string `json:"forbidden_tools,omitempty"`
	RequiredCaveats        []string `json:"required_caveats,omitempty"`
	ForbiddenClaims        []string `json:"forbidden_claims,omitempty"`
	RequireNumericFactRefs bool     `json:"require_numeric_fact_refs,omitempty"`
}

type NumericClaim struct {
	Text   string
	FactID string
}

// Trace is emitted by the agent runtime around a response. ClaimClasses are
// deterministic policy classifications, never free-form judge prose.
type Trace struct {
	ToolCalls     []string
	CaveatCodes   []string
	ClaimClasses  []string
	NumericClaims []NumericClaim
}

type Failure struct {
	Code, Detail string
}

func Load() (Corpus, error) {
	var corpus Corpus
	if err := json.Unmarshal(corpusData, &corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode evaluation corpus: %w", err)
	}
	if corpus.Schema != Schema || len(corpus.Cases) == 0 {
		return Corpus{}, fmt.Errorf("invalid evaluation corpus header")
	}
	seen := make(map[string]bool)
	workflows := make(map[string]bool)
	for _, testCase := range corpus.Cases {
		if testCase.ID == "" || testCase.Prompt == "" || seen[testCase.ID] {
			return Corpus{}, fmt.Errorf("invalid or duplicate evaluation case %q", testCase.ID)
		}
		seen[testCase.ID] = true
		workflows[testCase.Workflow] = true
	}
	for _, workflow := range []string{"explore", "experiment", "interpret"} {
		if !workflows[workflow] {
			return Corpus{}, fmt.Errorf("evaluation corpus has no %s case", workflow)
		}
	}
	return corpus, nil
}

func Evaluate(testCase Case, trace Trace) []Failure {
	var failures []Failure
	for _, required := range testCase.RequiredTools {
		if !slices.Contains(trace.ToolCalls, required) {
			failures = append(failures, Failure{Code: "missing_tool", Detail: required})
		}
	}
	for _, forbidden := range testCase.ForbiddenTools {
		if slices.Contains(trace.ToolCalls, forbidden) {
			failures = append(failures, Failure{Code: "forbidden_tool", Detail: forbidden})
		}
	}
	for _, required := range testCase.RequiredCaveats {
		if !slices.Contains(trace.CaveatCodes, required) {
			failures = append(failures, Failure{Code: "missing_caveat", Detail: required})
		}
	}
	for _, forbidden := range testCase.ForbiddenClaims {
		if slices.Contains(trace.ClaimClasses, forbidden) {
			failures = append(failures, Failure{Code: "forbidden_claim", Detail: forbidden})
		}
	}
	if testCase.RequireNumericFactRefs {
		if len(trace.NumericClaims) == 0 {
			failures = append(failures, Failure{Code: "missing_numeric_claim", Detail: "expected at least one traceable numeric claim"})
		}
		for _, claim := range trace.NumericClaims {
			if claim.FactID == "" {
				failures = append(failures, Failure{Code: "untraceable_numeric_claim", Detail: claim.Text})
			}
		}
	}
	return failures
}
