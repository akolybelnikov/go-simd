// Package modelcatalog describes evidence for model/workflow combinations.
// Availability is discovered separately; installation is not a recommendation.
package modelcatalog

import (
	"slices"
	"strings"
)

const EvidenceURL = "https://github.com/akolybelnikov/go-simd/tree/main/docs/evaluations/2026-11-22"

type Assessment struct {
	Status      string `json:"status"`
	Note        string `json:"note"`
	Recommended bool   `json:"recommended"`
	Evaluated   string `json:"evaluated,omitempty"`
	EvidenceURL string `json:"evidence_url,omitempty"`
}

// Recommendations are scoped to browser chat. A narrow CLI success is not enough.
var browserAssessments = map[string]Assessment{
	"qwen3:4b":       {Note: "Testing found unsolicited comparisons, incorrect explanations, and timeouts.", Evaluated: "2026-11-22"},
	"qwen3.5:4b":     {Note: "Testing found incorrect operation selection and missing comparison drafts.", Evaluated: "2026-11-22"},
	"granite4.2:3b":  {Note: "Testing found incorrect data types and claims about unmeasured array sizes.", Evaluated: "2026-11-22"},
	"qwen3.5:9b":     {Note: "Browser planning failed. Two CLI tool/citation checks passed, which is not a browser-chat recommendation.", Evaluated: "2026-11-22"},
	"inkling-2-mini": {Note: "Reviewed but not run: the published checkpoint exceeded our test host's memory. No browser-chat recommendation."},
}

func Browser(model string) Assessment {
	name := strings.TrimSuffix(model, ":latest")
	if name == "inkling-2-mini:26b-a5b" || name == "inkling-2-mini:26b-a5b-nvfp4" {
		name = "inkling-2-mini"
	}
	a, found := browserAssessments[name]
	a.Status = "Not evaluated"
	if !found {
		a.Note = "This installed model has not been evaluated for SIMD Lab browser chat."
		return a
	}
	a.EvidenceURL = EvidenceURL
	if a.Evaluated != "" {
		a.Status = "Evaluated — experimental"
	}
	if a.Recommended {
		a.Status = "Recommended for browser chat"
	}
	return a
}

func BrowserRecommendations() []string {
	var names []string
	for name, a := range browserAssessments {
		if a.Recommended {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

func BrowserNotice() string {
	if len(BrowserRecommendations()) > 0 {
		return "Local model quality depends on the model and task. Choose a model with a browser-chat recommendation and review its tested settings in the setup guide."
	}
	return "Local Ollama chat is experimental. The models tested so far made mistakes when choosing operations and preparing comparisons. We do not yet recommend a local model for browser chat."
}
