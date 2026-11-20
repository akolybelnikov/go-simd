package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labdocs"
)

const (
	chatQuestionBytes   = 4096
	chatReplyBytes      = 8192
	chatHistoryMessages = 6
	chatSessionLimit    = 24
)

type chatProviderOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Model string `json:"model"`
}

type chatSelection struct {
	ResultRef string `json:"result_ref,omitempty"`
	Size      int    `json:"size,omitempty"`
	Group     string `json:"group,omitempty"`
}

type chatCapture struct {
	ref, group, fixture, benchtime string
	inspection                     analysis.Interpretation
}

type chatDraft struct {
	Group   string `json:"group"`
	Method  string `json:"method"`
	Size    int    `json:"size"`
	Repeats int    `json:"repeats"`
}

type chatAnswer struct {
	Answer     string     `json:"answer"`
	Hypothesis string     `json:"hypothesis"`
	FactIDs    []string   `json:"fact_ids"`
	Draft      *chatDraft `json:"draft"`
}

type chatFact struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type chatResultContext struct {
	Group     string                   `json:"group"`
	Fixture   string                   `json:"fixture"`
	Benchtime string                   `json:"measurement_target_per_repeat"`
	Status    string                   `json:"status"`
	Size      int                      `json:"values_per_array"`
	Host      analysis.EnvironmentFact `json:"measuring_process"`
	Facts     []chatFact               `json:"facts"`
	Caveats   []analysis.Caveat        `json:"caveats"`
}

type chatContext struct {
	Catalog       []chatOperation    `json:"catalog"`
	SelectedGroup string             `json:"selected_group,omitempty"`
	Result        *chatResultContext `json:"result,omitempty"`
	Notes         []labdocs.Result   `json:"wiki_notes,omitempty"`
	MaxRepeats    int                `json:"max_repeats"`
}

type chatOperation struct {
	Group      string                      `json:"group"`
	Title      string                      `json:"title"`
	InputType  string                      `json:"input_type"`
	Output     labapi.ValueDescriptor      `json:"output"`
	Expression string                      `json:"expression"`
	Policy     string                      `json:"numeric_rules"`
	Methods    []labapi.StrategyDescriptor `json:"methods"`
	Baseline   string                      `json:"reference_method"`
	MaxSize    int                         `json:"max_size"`
}

func (app *localUI) rememberResult(capture evidence.Capture, inspection analysis.Interpretation, view uiResult) uiResult {
	snapshot := &chatCapture{rand.Text(), capture.Group, view.Fixture, view.Benchtime, inspection}
	app.mu.Lock()
	app.chatCapture = snapshot
	app.mu.Unlock()
	view.ChatRef = snapshot.ref
	return view
}

func (app *localUI) chatOptions() []chatProviderOption {
	options := make([]chatProviderOption, 0, len(app.chatProviders))
	for _, provider := range app.chatProviders {
		options = append(options, chatProviderOption{provider.id, provider.label, provider.model})
	}
	return options
}

func (app *localUI) buildChatContext(selection chatSelection, question string) (chatContext, error) {
	selected := findGuideOperation(app.catalog.Operations, selection.Group)
	if selection.Group != "" && selected == nil {
		return chatContext{}, fmt.Errorf("Open a supported operation to start a new chat.")
	}
	notes, _ := labdocs.Search(shortChatText(question, 256))
	data := chatContext{
		SelectedGroup: selection.Group, Notes: notes,
		MaxRepeats: labapi.DefaultPlanPolicy().MaxCount,
	}
	if selection.ResultRef == "" && selection.Size == 0 {
		data.Catalog = app.chatCatalog("")
		return data, nil
	}

	app.mu.Lock()
	snapshot := app.chatCapture
	app.mu.Unlock()
	if snapshot == nil || selection.ResultRef != snapshot.ref || selection.Size < 1 {
		return chatContext{}, fmt.Errorf("This result is no longer available to chat. Reopen the capture and start a new chat.")
	}
	inspection := resultAtSize(snapshot.inspection, selection.Size)
	if len(inspection.Series) == 0 {
		return chatContext{}, fmt.Errorf("Choose a measured batch size and start a new chat.")
	}
	host := inspection.MeasuringEnvironment
	host.CPU = shortChatText(host.CPU, 200)
	host.GoVersion = shortChatText(host.GoVersion, 100)
	host.OS = shortChatText(host.OS, 40)
	host.Arch = shortChatText(host.Arch, 40)
	result := &chatResultContext{
		Group: snapshot.group, Fixture: shortChatText(snapshot.fixture, 150), Benchtime: shortChatText(snapshot.benchtime, 40),
		Status: inspection.Status, Size: selection.Size, Host: host, Caveats: inspection.Caveats,
	}
	for _, series := range inspection.Series {
		result.Facts = append(result.Facts, chatFact{series.FactID, fmt.Sprintf(
			"%s (%s) had a median of %.5g ns per batch, range %.5g–%.5g ns over %d repeats.",
			seriesMethodName(series), series.ExecutionMode, series.MedianNS, series.MinimumNS, series.MaximumNS, series.Samples)})
	}
	for _, comparison := range inspection.Comparisons {
		base := methodName(snapshot.group, comparison.BaselineStrategy, comparison.BaselineStrategy)
		candidate := methodName(snapshot.group, comparison.CandidateStrategy, comparison.CandidateStrategy)
		result.Facts = append(result.Facts, chatFact{comparison.FactID, fmt.Sprintf(
			"%s median time was %+.4g%% relative to %s (baseline/candidate time ratio %.5g).",
			candidate, comparison.TimeChangePercent, base, comparison.SpeedupRatio)})
	}
	vectorMode := "native SIMD"
	if host.Emulated {
		vectorMode = "emulated SIMD"
	}
	result.Facts = append(result.Facts, chatFact{"measuring-process", fmt.Sprintf(
		"Measuring process was %s on %s/%s with %d-bit vectors, %s, and GOMAXPROCS=%d.",
		host.CPU, host.OS, host.Arch, host.VectorBits, vectorMode, host.GOMAXPROCS)})
	data.Result = result
	data.Catalog = app.chatCatalog(snapshot.group)
	return data, nil
}

func (app *localUI) chatCatalog(group string) []chatOperation {
	var choices []chatOperation
	for _, operation := range app.catalog.Operations {
		if group != "" && operation.Group != group {
			continue
		}
		choices = append(choices, chatOperation{
			Group: operation.Group, Title: operation.Title, InputType: operation.TypeID, Output: operation.Output,
			Expression: operation.Expression, Policy: operation.NumericalPolicy, Methods: operation.Strategies,
			Baseline: operation.ReferenceStrategy, MaxSize: operation.MaxSize,
		})
	}
	return choices
}

func shortChatText(value string, limit int) string {
	value = terminalText(value)
	if len(value) > limit {
		return strings.ToValidUTF8(value[:limit], "")
	}
	return value
}

type uiChatRequest struct {
	Provider  string        `json:"provider,omitempty"`
	Selection chatSelection `json:"selection"`
	Question  string        `json:"question"`
	History   []chatMessage `json:"history,omitempty"`
}

func (app *localUI) chat(w http.ResponseWriter, r *http.Request, preview bool) {
	var request uiChatRequest
	if err := decodeBounded(r.Body, 128<<10, &request); err != nil || !request.valid() {
		uiError(w, http.StatusBadRequest, "Use a short question and up to three recent exchanges.")
		return
	}
	labContext, err := app.buildChatContext(request.Selection, request.Question)
	if err != nil {
		uiError(w, http.StatusConflict, err.Error())
		return
	}
	if preview {
		uiJSON(w, http.StatusOK, labContext)
		return
	}
	providerIndex := slices.IndexFunc(app.chatProviders, func(p chatProvider) bool { return p.id == request.Provider })
	if providerIndex < 0 || strings.TrimSpace(request.Question) == "" {
		uiError(w, http.StatusBadRequest, "Choose a configured provider and enter a question.")
		return
	}
	provider := app.chatProviders[providerIndex]
	if status, message := app.beginModel(provider.id); status != 0 {
		uiError(w, status, message)
		return
	}
	defer app.endModel()

	contextJSON, err := json.Marshal(labContext)
	if err != nil || len(contextJSON) > 48<<10 {
		uiError(w, http.StatusUnprocessableEntity, "This context is too large. Ask about the catalog or choose a simpler capture.")
		return
	}
	ctx, cancel := context.WithDeadline(r.Context(), app.expires)
	defer cancel()
	answer, err := provider.answer(ctx, request.History, request.Question, contextJSON)
	if err != nil {
		uiError(w, http.StatusBadGateway, provider.label+": "+err.Error()+". You can also copy the question.")
		return
	}
	if request.Selection.ResultRef != "" {
		app.mu.Lock()
		current := app.chatCapture != nil && app.chatCapture.ref == request.Selection.ResultRef
		app.mu.Unlock()
		if !current {
			uiError(w, http.StatusConflict, "The selected result changed. Open it and start a new chat.")
			return
		}
	}
	var reply uiChatReply
	if provider.id == "openai" {
		reply, err = app.checkedChatAnswer(answer, labContext.Result)
	} else {
		reply, err = app.ollamaChatReply(answer, labContext.Result)
		if err != nil && labContext.Result != nil {
			reply = uiChatReply{
				Answer:  "Ollama returned an unreadable reply. Here are the measurements for the selected batch. You can retry or use Copy to another assistant.",
				LabOnly: true, Facts: labContext.Result.Facts, Caveats: labContext.Result.Caveats,
			}
			err = nil
		}
	}
	if err != nil {
		uiError(w, http.StatusBadGateway, provider.label+" returned an unreadable reply. Retry or use Copy to another assistant.")
		return
	}
	uiJSON(w, http.StatusOK, reply)
}

func (r uiChatRequest) valid() bool {
	if len(r.Question) > chatQuestionBytes || len(r.History) > chatHistoryMessages || len(r.History)%2 != 0 {
		return false
	}
	for index, message := range r.History {
		role := "user"
		limit := chatQuestionBytes
		if index%2 != 0 {
			role = "assistant"
			limit = chatReplyBytes
		}
		if message.Role != role || len(message.Content) > limit || strings.TrimSpace(message.Content) == "" {
			return false
		}
	}
	return true
}

type uiChatReply struct {
	Answer     string            `json:"answer"`
	LabOnly    bool              `json:"lab_only,omitempty"`
	Hypothesis string            `json:"hypothesis,omitempty"`
	Facts      []chatFact        `json:"facts,omitempty"`
	Caveats    []analysis.Caveat `json:"caveats,omitempty"`
	Draft      *chatDraft        `json:"draft,omitempty"`
	DraftError string            `json:"draft_error,omitempty"`
}

func (app *localUI) checkedChatAnswer(raw string, result *chatResultContext) (uiChatReply, error) {
	var answer chatAnswer
	if err := decodeBounded(strings.NewReader(raw), 16<<10, &answer); err != nil {
		return uiChatReply{}, err
	}
	if strings.TrimSpace(answer.Answer) == "" || len(answer.Answer) > chatQuestionBytes || len(answer.Hypothesis) > 2000 || len(answer.FactIDs) > 12 {
		return uiChatReply{}, fmt.Errorf("missing or excessive answer content")
	}
	reply := uiChatReply{Answer: answer.Answer, Hypothesis: answer.Hypothesis}
	if result != nil {
		reply.Caveats = result.Caveats
	}
	for _, id := range answer.FactIDs {
		if result == nil {
			return uiChatReply{}, fmt.Errorf("fact without result")
		}
		index := slices.IndexFunc(result.Facts, func(fact chatFact) bool { return fact.ID == id })
		if index < 0 {
			return uiChatReply{}, fmt.Errorf("fact outside selected result")
		}
		if !slices.Contains(reply.Facts, result.Facts[index]) {
			reply.Facts = append(reply.Facts, result.Facts[index])
		}
	}
	if answer.Draft != nil {
		app.attachChatDraft(&reply, *answer.Draft, result)
	}
	return reply, nil
}

const chatDraftUnavailable = "The suggested settings are unavailable. You can set up a comparison in Explore."

// Ollama models can add fields or wrap a JSON object in a Markdown fence.
// Only recognized fields get through, and all displayed evidence comes from Go.
func (app *localUI) ollamaChatReply(raw string, result *chatResultContext) (uiChatReply, error) {
	if len(raw) > 16<<10 {
		return uiChatReply{}, fmt.Errorf("Ollama reply exceeds response limit")
	}
	raw = strings.TrimSpace(raw)
	if fence, body, ok := strings.Cut(raw, "\n"); ok &&
		(strings.TrimSpace(fence) == "```json" || strings.TrimSpace(fence) == "```") {
		if content, ok := strings.CutSuffix(body, "\n```"); ok {
			raw = content
		}
	}
	var fields struct {
		Answer     string          `json:"answer"`
		Hypothesis json.RawMessage `json:"hypothesis"`
		Draft      json.RawMessage `json:"draft"`
	}
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return uiChatReply{}, err
	}
	if strings.TrimSpace(fields.Answer) == "" || len(fields.Answer) > chatQuestionBytes {
		return uiChatReply{}, fmt.Errorf("missing or excessive Ollama answer content")
	}
	reply := uiChatReply{Answer: fields.Answer}
	var hypothesis string
	if json.Unmarshal(fields.Hypothesis, &hypothesis) == nil && len(hypothesis) <= 2000 {
		reply.Hypothesis = hypothesis
	}
	if result != nil {
		reply.Facts, reply.Caveats = result.Facts, result.Caveats
	}
	if len(fields.Draft) > 0 {
		var draft *chatDraft
		if err := decodeBounded(strings.NewReader(string(fields.Draft)), 16<<10, &draft); err != nil {
			reply.DraftError = chatDraftUnavailable
		} else if draft != nil {
			app.attachChatDraft(&reply, *draft, result)
		}
	}
	return reply, nil
}

func (app *localUI) attachChatDraft(reply *uiChatReply, draft chatDraft, result *chatResultContext) {
	if app.validChatDraft(draft, result) {
		reply.Draft = &draft
	} else {
		reply.DraftError = chatDraftUnavailable
	}
}

func (app *localUI) validChatDraft(draft chatDraft, result *chatResultContext) bool {
	if result != nil && result.Group != draft.Group {
		return false
	}
	operation := findGuideOperation(app.catalog.Operations, draft.Group)
	policy := labapi.DefaultPlanPolicy()
	if operation == nil || draft.Repeats < 1 || draft.Repeats > policy.MaxCount {
		return false
	}
	methods, err := guideMethods(*operation, []string{operation.ReferenceStrategy, draft.Method})
	if err != nil {
		return false
	}
	request := comparisonRequest(draft.Group, methods, draft.Size)
	request.Count = draft.Repeats
	_, err = labapi.ResolveExperimentPlan(request, app.catalog.Runtime, policy)
	return err == nil
}

func (app *localUI) beginModel(provider string) (int, string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.running || app.chatRunning {
		return http.StatusConflict, "Wait for the current comparison or chat request to finish."
	}
	if provider == "openai" {
		if app.hostedTurns >= chatSessionLimit {
			return http.StatusTooManyRequests, "This lab session has used its OpenAI turns. Use Ollama or copy the question."
		}
		if time.Since(app.lastHostedTurn) < 2*time.Second {
			return http.StatusTooManyRequests, "Wait a moment before sending another OpenAI question."
		}
		app.hostedTurns++
		app.lastHostedTurn = time.Now()
	}
	app.chatRunning = true
	return 0, ""
}

func (app *localUI) endModel() {
	app.mu.Lock()
	app.chatRunning = false
	app.mu.Unlock()
}
