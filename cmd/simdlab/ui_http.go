package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/akolybelnikov/go-simd/docs/wiki"
	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/analysis"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labservice"
	"github.com/akolybelnikov/go-simd/internal/labtools"
)

//go:embed ui/index.html ui/app.js ui/style.css
var uiAssets embed.FS

const uiSessionTTL = 30 * time.Minute
const uiModelTimeout = 30 * time.Second

type localUI struct {
	host, origin, token string
	expires             time.Time
	root                string
	model               agent.ModelClient
	catalog             labapi.CatalogResponse
	chatProviders       []chatProvider

	mu             sync.Mutex
	pending        *uiProposal
	running        bool
	chatCapture    *chatCapture
	chatRunning    bool
	hostedTurns    int
	lastHostedTurn time.Time
}

type uiProposal struct {
	ref     string
	plan    labapi.ExperimentPlan
	tools   *labtools.Toolset
	expires time.Time
}

type uiPreview struct {
	Ref       string   `json:"ref"`
	Operation string   `json:"operation"`
	Methods   []string `json:"methods"`
	Size      int      `json:"size"`
	Repeats   int      `json:"repeats"`
	Benchtime string   `json:"benchtime"`
	Host      string   `json:"host"`
	Warnings  []string `json:"warnings,omitempty"`
	TailNotes []string `json:"tail_notes,omitempty"`
	Ollama    bool     `json:"ollama"`
}

func newLocalUI(host, root string, service *labservice.Deterministic, model agent.ModelClient) (*localUI, error) {
	available, err := service.Catalog("all")
	if err != nil {
		return nil, err
	}
	return &localUI{
		host: host, origin: "http://" + host, token: rand.Text(), expires: time.Now().Add(uiSessionTTL),
		root: root, catalog: available, model: model,
	}, nil
}

func (app *localUI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")

	if r.Host != app.host {
		http.Error(w, "Invalid local address.", http.StatusForbidden)
		return
	}
	switch r.URL.Path {
	case "/", "/app.js", "/style.css":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Use GET.", http.StatusMethodNotAllowed)
			return
		}
		app.serveAsset(w, r.URL.Path)
		return
	}
	if r.URL.Path == "/wiki" || strings.HasPrefix(r.URL.Path, "/wiki/") {
		serveWiki(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		uiError(w, http.StatusMethodNotAllowed, "Use POST.")
		return
	}
	if r.Header.Get("Origin") != app.origin {
		uiError(w, http.StatusForbidden, "Open the exact local address printed by simdlab.")
		return
	}
	if !time.Now().Before(app.expires) || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Simdlab-Session")), []byte(app.token)) != 1 {
		uiError(w, http.StatusUnauthorized, "Open the private URL from simdlab. If the session expired, restart simdlab ui.")
		return
	}

	switch r.URL.Path {
	case "/api/catalog":
		app.explore(w)
	case "/api/preview":
		app.preview(w, r)
	case "/api/starter":
		app.propose(w, additionStarter())
	case "/api/run":
		app.run(w, r)
	case "/api/open":
		app.open(w, r)
	case "/api/ask":
		app.ask(w, r)
	case "/api/chat", "/api/chat/context":
		app.chat(w, r, r.URL.Path == "/api/chat/context")
	default:
		uiError(w, http.StatusNotFound, "Unknown lab action.")
	}
}

func (app *localUI) explore(w http.ResponseWriter) {
	type operationView struct {
		labapi.OperationDescriptor
		Summary  string `json:"summary"`
		GuideURL string `json:"guide_url"`
	}
	view := struct {
		labapi.CatalogResponse
		Operations     []operationView      `json:"operations"`
		DefaultRepeats int                  `json:"default_repeats"`
		MaxRepeats     int                  `json:"max_repeats"`
		ChatProviders  []chatProviderOption `json:"chat_providers"`
	}{
		CatalogResponse: app.catalog, DefaultRepeats: starterRepeats,
		MaxRepeats:    labapi.DefaultPlanPolicy().MaxCount,
		ChatProviders: app.chatOptions(),
	}
	for _, operation := range app.catalog.Operations {
		operation.Strategies = slices.Clone(operation.Strategies)
		for j := range operation.Strategies {
			method := &operation.Strategies[j]
			method.Label = methodName(operation.Group, method.ID, method.Label)
		}
		note := wiki.Operation(operation.Group)
		view.Operations = append(view.Operations, operationView{operation, note.Summary, note.URL})
	}
	uiJSON(w, http.StatusOK, view)
}

func (app *localUI) preview(w http.ResponseWriter, r *http.Request) {
	var choice struct {
		Group   string `json:"group"`
		Method  string `json:"method"`
		Size    int    `json:"size"`
		Repeats *int   `json:"repeats"`
	}
	if err := decodeBounded(r.Body, 1024, &choice); err != nil || choice.Group == "" || choice.Method == "" {
		uiError(w, http.StatusBadRequest, "Choose an operation, a method, and a whole number for the batch size.")
		return
	}
	operation := findGuideOperation(app.catalog.Operations, choice.Group)
	if operation == nil {
		uiError(w, http.StatusUnprocessableEntity, "Choose a supported operation from Explore.")
		return
	}
	repeats := starterRepeats
	if choice.Repeats != nil {
		repeats = *choice.Repeats
	}
	app.previewComparison(w, *operation, []string{operation.ReferenceStrategy, choice.Method}, choice.Size, repeats)
}

func (app *localUI) serveAsset(w http.ResponseWriter, path string) {
	name, contentType := "ui/index.html", "text/html; charset=utf-8"
	switch path {
	case "/app.js":
		name, contentType = "ui/app.js", "text/javascript; charset=utf-8"
	case "/style.css":
		name, contentType = "ui/style.css", "text/css; charset=utf-8"
	}
	content, err := uiAssets.ReadFile(name)
	if err != nil {
		http.Error(w, "The embedded lab interface is unavailable.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(content)
}

func (app *localUI) propose(w http.ResponseWriter, request labapi.PlanRequest) {
	tools, err := newComparisonTools(app.root)
	if err != nil {
		uiError(w, http.StatusInternalServerError, "Could not start the measuring worker.")
		return
	}
	plan, err := tools.ProposeExperiment(request)
	if err != nil {
		uiError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	proposal := &uiProposal{ref: rand.Text(), plan: plan, tools: tools, expires: time.Now().Add(labtools.ExperimentPlanTTL)}
	app.mu.Lock()
	if app.running {
		app.mu.Unlock()
		uiError(w, http.StatusConflict, "A comparison is already measuring. Wait for it to finish.")
		return
	}
	app.pending = proposal
	app.mu.Unlock()

	preview := uiPreview{
		Ref: proposal.ref, Operation: plan.Group, Repeats: plan.Count, Size: plan.Cases[0].Elements, Benchtime: plan.Benchtime,
		Host: fmt.Sprintf("%s / %s", plan.Runtime.OS, plan.Runtime.Arch), Ollama: app.model != nil,
		TailNotes: comparisonTailNotes(plan),
	}
	if operation := findGuideOperation(app.catalog.Operations, plan.Group); operation != nil {
		preview.Operation = operation.Title
	}
	for _, c := range plan.Cases {
		preview.Methods = append(preview.Methods, methodName(plan.Group, c.StrategyID, c.StrategyLabel))
	}
	for _, warning := range plan.Warnings {
		preview.Warnings = append(preview.Warnings, warning.Message)
	}
	uiJSON(w, http.StatusOK, preview)
}

func (app *localUI) run(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Ref string `json:"ref"`
	}
	if err := decodeBounded(r.Body, 1024, &request); err != nil {
		uiError(w, http.StatusBadRequest, "Send the reference for the plan you reviewed.")
		return
	}

	app.mu.Lock()
	if app.running {
		app.mu.Unlock()
		uiError(w, http.StatusConflict, "A comparison is already measuring.")
		return
	}
	if app.chatRunning {
		app.mu.Unlock()
		uiError(w, http.StatusConflict, "Wait for the model request to finish before measuring.")
		return
	}
	pending := app.pending
	if pending == nil || pending.ref != request.Ref || !time.Now().Before(pending.expires) {
		app.mu.Unlock()
		uiError(w, http.StatusConflict, "This preview is no longer available. Open the starter or ask for a new preview.")
		return
	}
	app.pending, app.running = nil, true
	app.mu.Unlock()
	defer func() {
		app.mu.Lock()
		app.running = false
		app.mu.Unlock()
	}()

	saved, err := executeComparison(r.Context(), pending.tools, pending.plan)
	if err != nil {
		if saved.path != "" && saved.inspection.Status == "" {
			err = fmt.Errorf("%w. Partial measurements: %s", err, saved.path)
		}
		uiError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	view := newUIResult(*saved.capture, saved.inspection, saved.report, saved.path, saved.reportPath)
	uiJSON(w, http.StatusOK, app.rememberResult(*saved.capture, saved.inspection, view))
}

func (app *localUI) open(w http.ResponseWriter, r *http.Request) {
	limits := evidence.DefaultReadLimits()
	if r.ContentLength > limits.MaxBytes {
		uiError(w, http.StatusRequestEntityTooLarge, "Choose a measurement capture smaller than 8 MiB.")
		return
	}
	capture, err := evidence.ReadCapture(http.MaxBytesReader(w, r.Body, limits.MaxBytes), limits)
	if err != nil {
		uiError(w, http.StatusUnprocessableEntity, "Could not open this capture. "+err.Error())
		return
	}
	inspection, err := analysis.InterpretCapture(capture)
	if err != nil {
		uiError(w, http.StatusUnprocessableEntity, "Could not compare this capture. "+err.Error())
		return
	}
	report := comparisonReport(capture, inspection, false)
	view := newUIResult(capture, inspection, report, "", "")
	uiJSON(w, http.StatusOK, app.rememberResult(capture, inspection, view))
}

func (app *localUI) ask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Question string `json:"question"`
	}
	if err := decodeBounded(r.Body, 6*maxQuestionBytes+256, &body); err != nil || len(body.Question) > maxQuestionBytes || strings.TrimSpace(body.Question) == "" {
		uiError(w, http.StatusBadRequest, "Enter a short question about an operation.")
		return
	}
	request, matched := matchGuideRequest(strings.TrimSpace(body.Question))
	if !matched {
		if app.model == nil {
			uiError(w, http.StatusUnprocessableEntity, "Try “Compare summing an array with a loop and with vectors.” For other wording, start simdlab ui -ollama.")
			return
		}
		if status, message := app.beginModel("ollama"); status != 0 {
			uiError(w, status, message)
			return
		}
		defer app.endModel()
		ctx, cancel := context.WithTimeout(r.Context(), uiModelTimeout)
		defer cancel()
		var err error
		request, err = routeGuideRequest(ctx, app.model, app.catalog.Operations, body.Question, "", false)
		if err != nil {
			uiError(w, http.StatusUnprocessableEntity, "Ollama could not resolve the question. Try wording from the examples.")
			return
		}
	}
	if request.Action == "unsupported" {
		uiError(w, http.StatusUnprocessableEntity, guideScope)
		return
	}
	if request.Action == "invalid_size" {
		uiError(w, http.StatusUnprocessableEntity, uiSizeRange())
		return
	}
	if request.Action != "compare" {
		uiError(w, http.StatusUnprocessableEntity, "Name one operation and the methods to compare. To see earlier timings, open a saved run.")
		return
	}
	operation := findGuideOperation(app.catalog.Operations, request.Group)
	if operation == nil {
		uiError(w, http.StatusUnprocessableEntity, "Name a catalog operation such as addition or array sum.")
		return
	}
	size := request.Size
	if size == 0 {
		size = starterSize
	}
	app.previewComparison(w, *operation, request.Methods, size, starterRepeats)
}

func (app *localUI) previewComparison(w http.ResponseWriter, operation labapi.OperationDescriptor, requested []string, size, repeats int) {
	methods, err := guideMethods(operation, requested)
	if err != nil {
		uiError(w, http.StatusUnprocessableEntity, "Choose two supported methods, including the ordinary scalar method.")
		return
	}
	if maxRepeats := labapi.DefaultPlanPolicy().MaxCount; repeats < 1 || repeats > maxRepeats {
		uiError(w, http.StatusUnprocessableEntity, fmt.Sprintf("Choose between 1 and %d repeats per method.", maxRepeats))
		return
	}
	request := comparisonRequest(operation.Group, methods, size)
	request.Count = repeats
	app.propose(w, request)
}

func uiSizeRange() string {
	return fmt.Sprintf("Choose a whole number from 1 to %d values per input array.", labapi.MaxExperimentSize)
}

func uiJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func uiError(w http.ResponseWriter, status int, message string) {
	uiJSON(w, status, map[string]string{"error": message})
}
