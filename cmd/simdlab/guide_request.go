package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/labapi"
)

type guideRequest struct {
	Action  string   `json:"action"`
	Group   string   `json:"group,omitempty"`
	Methods []string `json:"methods,omitempty"`
	Size    int      `json:"size,omitempty"`
}

var starterComparison = regexp.MustCompile(`^(?:please )?(?:compare|benchmark) (?:(?:ordinary|scalar) (?:and|vs\.?|versus) (?:vectorized|vectorised|simd) )?addition(?: (?:for|at|on|with) (?:(?:size|arrays of) )?([0-9,]+)(?: (?:numbers|elements|values)(?: per array)?)?)?$`)
var arrayAdditionComparison = regexp.MustCompile(`^(?:(?:please|can you|could you|can we) )?(?:compare|benchmark|time) (?:(?:two|different) ways of )?(?:adding|add) (?:two )?arrays(?: (?:of|with) ([0-9,]+) (?:values|elements|numbers)(?: each| per array)?)?(?: with (?:a |the )?(?:scalar )?loop (?:and|versus|vs\.?) (?:with )?(?:vectors|simd))?$`)
var arraySumComparison = regexp.MustCompile(`^(?:(?:please|can you|could you|can we) )?(?:compare|benchmark|time) (?:(?:two|different) ways of )?(?:summing|sum) (?:one|an) array(?: (?:of|with) ([0-9,]+) (?:values|elements|numbers))?(?: with (?:a |the )?(?:scalar )?loop (?:and|versus|vs\.?) (?:with )?(?:vectors|simd))?$`)
var sumDiscovery = regexp.MustCompile(`^(?:can i compare (?:different )?ways of summing (?:an|a float32) array|how can i compare ways of summing an array|what are the (?:ways|methods) (?:to sum|of summing) an array)$`)
var unsupportedScope = regexp.MustCompile(`\b(?:csv|parquet|excel|gpu|end[- ]to[- ]end|pipeline|(?:another|remote) (?:machine|computer|server|device)|my (?:data|file|dataset|laptop|phone|computer|server|machine))\b`)
var unsupportedInference = regexp.MustCompile(`\b(?:statistically significant|statistical significance|confidence interval)\b`)
var repeatQuestion = regexp.MustCompile(`^(?:were the (?:repeats|timings) consistent|how consistent were (?:the repeats|the timings)|(?:show|give me) (?:the )?(?:repeat|timing) ranges|what (?:was|were) the (?:repeat|timing) ranges)$`)
var resultQuestion = regexp.MustCompile(`^(?:which(?: method| one| version| approach| implementation)? (?:took less time|was faster|was (?:the )?fastest|ran faster|won|had the lower median)(?:,? and by how much)?|by how much|how much (?:faster|quicker) was (?:the )?(?:faster method|winner)|(?:explain|summarize|show) (?:the |my )?(?:saved |last )?(?:result|results))$`)
var resultQualifier = regexp.MustCompile(` (?:at|for|on|with) `)
var resultBatch = regexp.MustCompile(`^(?:(?:size|a batch of) )?([0-9,]+)(?: (?:numbers|elements|values)(?: per array)?)?$`)
var guideInteger = regexp.MustCompile(`^(?:[1-9][0-9]*|[1-9][0-9]{0,2}(?:,[0-9]{3})+)$`)

func matchGuideRequest(question string) (guideRequest, bool) {
	normalized := strings.Join(strings.Fields(strings.ToLower(strings.TrimRight(question, ".?!"))), " ")
	switch {
	case unsupportedScope.MatchString(normalized):
		return guideRequest{Action: "unsupported"}, true
	case unsupportedInference.MatchString(normalized):
		return guideRequest{Action: "inference"}, true
	case sumDiscovery.MatchString(normalized):
		return guideRequest{Action: "find", Group: "sum"}, true
	case normalized == "details":
		return guideRequest{Action: "details"}, true
	}
	if first, second, ok := strings.Cut(normalized, "? "); ok && repeatQuestion.MatchString(second) {
		if request, ok := matchResultQuestion(first); ok && request.Action == "result" {
			request.Action = "result_repeats"
			return request, true
		}
	}
	if request, ok := matchResultQuestion(normalized); ok {
		return request, true
	}
	for _, pattern := range []struct {
		expression *regexp.Regexp
		group      string
	}{
		{starterComparison, "addition"},
		{arrayAdditionComparison, "addition"},
		{arraySumComparison, "sum"},
	} {
		if match := pattern.expression.FindStringSubmatch(normalized); match != nil {
			size := starterSize
			if match[1] != "" {
				var ok bool
				size, ok = guideSize(match[1])
				if !ok {
					return guideRequest{Action: "invalid_size"}, true
				}
			}
			return guideRequest{Action: "compare", Group: pattern.group, Size: size}, true
		}
	}
	return guideRequest{}, false
}

func matchResultQuestion(question string) (guideRequest, bool) {
	question = strings.TrimSuffix(strings.TrimSuffix(question, " and by how much"), ",")
	if prefix, rest, ok := strings.Cut(question, ", "); ok {
		for _, preposition := range []string{"at ", "for "} {
			if qualifier, ok := strings.CutPrefix(prefix, preposition); ok {
				question = rest + " at " + qualifier
				break
			}
		}
	}
	parts := resultQualifier.Split(question, -1)
	var request guideRequest
	switch {
	case resultQuestion.MatchString(parts[0]):
		request.Action = "result"
	case repeatQuestion.MatchString(parts[0]):
		request.Action = "repeats"
	default:
		return guideRequest{}, false
	}
	for _, part := range parts[1:] {
		qualifier := strings.TrimPrefix(part, "the ")
		if match := resultBatch.FindStringSubmatch(qualifier); match != nil && request.Size == 0 {
			if size, ok := guideSize(match[1]); ok {
				request.Size = size
				continue
			}
		}
		if group := resultOperation(qualifier); group != "" && request.Group == "" {
			request.Group = group
			continue
		}
		// Unrecognized or multiple qualifiers must never select the whole run.
		return guideRequest{Action: "result_target"}, true
	}
	return request, true
}

func resultOperation(qualifier string) string {
	switch qualifier {
	case "adding arrays", "adding two arrays", "elementwise addition", "float32 addition":
		return "addition"
	case "float64 addition":
		return "addition64"
	case "summing an array", "summing one array", "array sum", "float32 sum":
		return "sum"
	case "dot product":
		return "dot"
	}
	if op, ok := catalog.ByGroup(qualifier); ok {
		return op.Group
	}
	return ""
}

func guideSize(text string) (int, bool) {
	if !guideInteger.MatchString(text) {
		return 0, false
	}
	size, err := strconv.Atoi(strings.ReplaceAll(text, ",", ""))
	return size, err == nil
}

const guideToolName = "route_lab_request"

const guideInstructions = `Map the user's question to one action by calling route_lab_request once.
Only select listed catalog groups and method IDs.
- find: select the operation the user wants to discover.
- compare: select one operation and one batch size. Any whole number from 1 to that operation's max_size is allowed. Include the exact size if specified, even if out of range. Leave size absent for the default of 1024. Omit methods for ordinary scalar and default serial SIMD, or supply two requested catalog IDs including scalar.
- clarify: ask which operation when multiple choices fit. Adding two arrays is addition; adding up one array is sum.
- unsupported: custom files, other hosts, pipelines, claims beyond timings, custom fixtures, tuning, multi-size sweeps, or anything outside this tool.
- result, repeats, result_repeats, details: the user is asking about measurements rather than requesting a new experiment. Do not propose a benchmark in response to a question about measurements.
Treat previous_question as context for a clarification. You cannot approve or run a benchmark.`

func routeGuideRequest(ctx context.Context, client agent.ModelClient, operations []labapi.OperationDescriptor, question, previous string, hasResults bool) (guideRequest, error) {
	type operationChoice struct {
		Group   string            `json:"group"`
		Title   string            `json:"title"`
		Methods map[string]string `json:"methods"`
		MaxSize int               `json:"max_size"`
	}
	choices := make([]operationChoice, 0, len(operations))
	for _, operation := range operations {
		methods := make(map[string]string, len(operation.Strategies))
		for _, method := range operation.Strategies {
			methods[method.ID] = methodName(operation.Group, method.ID, method.Label)
		}
		choices = append(choices, operationChoice{Group: operation.Group, Title: operation.Title, Methods: methods, MaxSize: operation.MaxSize})
	}
	data, err := json.Marshal(struct {
		Question         string            `json:"question"`
		PreviousQuestion string            `json:"previous_question,omitempty"`
		HasResults       bool              `json:"has_saved_results"`
		Operations       []operationChoice `json:"operations"`
	}{question, previous, hasResults, choices})
	if err != nil {
		return guideRequest{}, err
	}
	runtime, err := agent.NewRuntime(client, guideInstructions, []agent.ToolDefinition{{
		Name: guideToolName, Description: "Identify one lab action.",
		InputSchema: []byte(`{"type":"object","additionalProperties":false,"required":["action"],"properties":{"action":{"type":"string","enum":["find","compare","clarify","result","repeats","result_repeats","details","unsupported"]},"group":{"type":"string"},"methods":{"type":"array","maxItems":2,"items":{"type":"string"}},"size":{"type":"integer","minimum":1}}}`),
	}})
	if err != nil {
		return guideRequest{}, err
	}
	reply, err := runtime.RunModelStep(ctx, []agent.Message{{Role: agent.RoleUser, Content: string(data)}})
	if err != nil {
		return guideRequest{}, err
	}
	if len(reply.ToolCalls) != 1 || reply.ToolCalls[0].Name != guideToolName {
		return guideRequest{}, errors.New("expected exactly one lab request")
	}
	var request guideRequest
	if err := decodeBounded(bytes.NewReader(reply.ToolCalls[0].Arguments), maxQuestionBytes, &request); err != nil {
		return guideRequest{}, err
	}
	switch request.Action {
	case "compare":
		if request.Size < 0 || len(request.Methods) > 2 {
			return guideRequest{}, errors.New("invalid comparison")
		}
	case "find", "clarify", "result", "repeats", "result_repeats", "details", "unsupported":
	default:
		return guideRequest{}, errors.New("unknown lab action")
	}
	return request, nil
}
