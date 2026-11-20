package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/labservice"
	"github.com/akolybelnikov/go-simd/internal/labtools"
	"github.com/akolybelnikov/go-simd/internal/labworker"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/stretchr/testify/require"
)

func TestLocalUIRequiresSessionAndOneApproval(t *testing.T) {
	app := testLocalUI(t)
	page := httptest.NewRequest(http.MethodGet, app.origin+"/", nil)
	html := httptest.NewRecorder()
	app.ServeHTTP(html, page)
	require.Equal(t, http.StatusOK, html.Code)
	require.Contains(t, html.Body.String(), "Run locally")
	require.Contains(t, html.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
	require.NotContains(t, html.Body.String(), app.token)

	for _, tc := range []struct {
		name, host, origin, token string
		want                      int
	}{
		{"missing token", app.host, app.origin, "", http.StatusUnauthorized},
		{"wrong origin", app.host, "https://example.test", app.token, http.StatusForbidden},
		{"wrong host", "lab.example.test", app.origin, app.token, http.StatusForbidden},
		{"missing origin", app.host, "", app.token, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, app.origin+"/api/starter", nil)
			request.Host = tc.host
			request.Header.Set("Origin", tc.origin)
			request.Header.Set("X-Simdlab-Session", tc.token)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, request)
			require.Equal(t, tc.want, response.Code)
		})
	}

	previous := getUIStarter(t, app)
	current := getUIStarter(t, app)
	require.NotEqual(t, previous.Ref, current.Ref)
	require.Equal(t, "Addition", current.Operation)
	require.Equal(t, starterSize, current.Size)
	require.Equal(t, 3, current.Repeats)
	require.Len(t, current.Methods, 2)
	require.Equal(t, http.StatusConflict, postUI(app, "/api/run", approvalJSON(t, previous.Ref)).Code)
	entries, err := os.ReadDir(app.root)
	require.NoError(t, err)
	require.Empty(t, entries)

	response := postUI(app, "/api/run", approvalJSON(t, current.Ref))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result uiResult
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Contains(t, result.Timing, "Median per batch")
	require.Len(t, result.Series, 2)
	for _, series := range result.Series {
		require.Len(t, series.Repeats, current.Repeats)
	}
	report, err := os.ReadFile(result.ReportPath)
	require.NoError(t, err)
	require.Equal(t, result.Report, string(report))
	require.FileExists(t, result.Evidence)
	require.Equal(t, http.StatusConflict, postUI(app, "/api/run", approvalJSON(t, current.Ref)).Code)

	app.expires = time.Now().Add(-time.Second)
	require.Equal(t, http.StatusUnauthorized, postUI(app, "/api/starter", "").Code)
}

func TestLocalUICaptureAndOfflineQuestion(t *testing.T) {
	app := testLocalUI(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "features", "evidence", "sum-cancellation.measurements.json"))
	require.NoError(t, err)
	opened := postUI(app, "/api/open", string(data))
	require.Equal(t, http.StatusOK, opened.Code, opened.Body.String())
	var result uiResult
	require.NoError(t, json.Unmarshal(opened.Body.Bytes(), &result))
	require.Contains(t, result.Timing, "40% lower median in this run")
	require.NotEmpty(t, result.Host.CPU)
	require.Len(t, result.Series[0].Repeats, 2)
	require.Contains(t, strings.Join(result.Caveats, "\n"), "change the rounded result")
	require.Contains(t, result.Report, "Measuring host")
	require.Empty(t, result.ReportPath)
	require.Equal(t, http.StatusUnprocessableEntity, postUI(app, "/api/open", "{}").Code)
	oversized := postUI(app, "/api/open", strings.Repeat(" ", (8<<20)+1))
	require.Equal(t, http.StatusRequestEntityTooLarge, oversized.Code)

	starter := getUIStarter(t, app)
	invalid := postUI(app, "/api/ask", `{"question":"Compare addition for 16777217 values."}`)
	require.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	require.Contains(t, invalid.Body.String(), "Choose a whole number from 1 to 16777216")
	require.Equal(t, starter.Ref, app.pending.ref)
	response := postUI(app, "/api/ask", `{"question":"Compare summing an array with a loop and with vectors."}`)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var preview uiPreview
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	require.False(t, preview.Ollama)
	require.Equal(t, []string{"Ordinary sum (left to right)", "Vectorized sum (one accumulator)"}, preview.Methods)
	require.Equal(t, http.StatusConflict, postUI(app, "/api/run", approvalJSON(t, starter.Ref)).Code)
	entries, err := os.ReadDir(app.root)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestLocalUIExploreUsesCatalogMethodsAndSizes(t *testing.T) {
	app := testLocalUI(t)
	response := postUI(app, "/api/catalog", "")
	require.Equal(t, http.StatusOK, response.Code)
	var catalog labapi.CatalogResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &catalog))
	require.Equal(t, app.catalog.Runtime, catalog.Runtime)
	require.Len(t, catalog.Operations, len(app.catalog.Operations))
	sum := findGuideOperation(catalog.Operations, "sum")
	require.NotNil(t, sum)
	require.Contains(t, sum.SupportedSizes, 64)
	require.Nil(t, app.pending)

	for _, invalid := range []string{
		`{"group":"unknown","method":"simd-striped-2","size":64}`,
		`{"group":"sum","method":"simd-striped-2","size":0}`,
		`{"group":"sum","method":"scalar","size":64}`,
	} {
		require.Equal(t, http.StatusUnprocessableEntity, postUI(app, "/api/preview", invalid).Code)
		require.Nil(t, app.pending)
	}
	response = postUI(app, "/api/preview", `{"group":"sum","method":"simd-striped-2","size":1003}`)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var preview uiPreview
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	require.Equal(t, 1003, preview.Size)
	require.Equal(t, []string{"Ordinary sum (left to right)", "Vectorized sum (two accumulators)"}, preview.Methods)
	require.Equal(t, preview.Ref, app.pending.ref)
	entries, err := os.ReadDir(app.root)
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestLocalUIOperationGuidesResolveOffline(t *testing.T) {
	app := testLocalUI(t)
	catalog := postUI(app, "/api/catalog", "")
	require.Equal(t, http.StatusOK, catalog.Code)
	var response struct {
		Operations []struct {
			Group    string `json:"group"`
			Summary  string `json:"summary"`
			GuideURL string `json:"guide_url"`
		}
	}
	require.NoError(t, json.Unmarshal(catalog.Body.Bytes(), &response))
	require.Len(t, response.Operations, len(app.catalog.Operations))

	get := func(path string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, app.origin+path, nil)
		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)
		return recorder
	}
	home := get("/wiki/")
	require.Equal(t, http.StatusOK, home.Code)
	require.Contains(t, home.Body.String(), "<table>")
	require.Contains(t, home.Body.String(), `href="/wiki/04-Operation-guide"`)
	require.Contains(t, home.Body.String(), `href="/wiki/02-First-experiment"`)
	require.Equal(t, http.StatusNotFound, get("/wiki/../../go.mod").Code)

	pages := make(map[string]string)
	for _, operation := range response.Operations {
		t.Run(operation.Group, func(t *testing.T) {
			require.NotEmpty(t, operation.Summary)
			path, anchor, ok := strings.Cut(operation.GuideURL, "#")
			require.True(t, ok)
			require.NotEmpty(t, anchor)
			if pages[path] == "" {
				guide := get(path)
				require.Equal(t, http.StatusOK, guide.Code)
				pages[path] = guide.Body.String()
			}
			require.Contains(t, pages[path], `id="`+anchor+`"`)
			require.Contains(t, pages[path], "What is timed.")
		})
	}
}

func TestLocalUIRunsCustomAdditionSizes(t *testing.T) {
	app := testLocalUI(t)
	aligned := postUI(app, "/api/ask", `{"question":"Compare addition for 80,000 values."}`)
	require.Equal(t, http.StatusOK, aligned.Code, aligned.Body.String())
	var preview uiPreview
	require.NoError(t, json.Unmarshal(aligned.Body.Bytes(), &preview))
	require.Equal(t, 80000, preview.Size)
	require.Empty(t, preview.TailNotes)

	partial := postUI(app, "/api/ask", `{"question":"Compare addition for 80,003 values."}`)
	require.Equal(t, http.StatusOK, partial.Code, partial.Body.String())
	require.Contains(t, partial.Body.String(), "final partial vector handles 3 remaining values")

	scalarTail := postUI(app, "/api/preview", `{"group":"addition","method":"simd-scalar-tail","size":80003}`)
	require.Equal(t, http.StatusOK, scalarTail.Code, scalarTail.Body.String())
	require.NoError(t, json.Unmarshal(scalarTail.Body.Bytes(), &preview))
	require.Equal(t, 80003, preview.Size)
	require.Contains(t, preview.TailNotes[0], "scalar tail handles the last 3")

	response := postUI(app, "/api/run", approvalJSON(t, preview.Ref))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result uiResult
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Len(t, result.Series, 2)
	for _, series := range result.Series {
		require.Equal(t, 80003, series.Size)
		require.Len(t, series.Repeats, preview.Repeats)
	}
	require.FileExists(t, result.Evidence)
}

func TestLocalUIModelOnlyProposes(t *testing.T) {
	app := testLocalUI(t)
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = io.WriteString(w, `{"message":{"content":"Run it now.","tool_calls":[{"function":{"name":"route_lab_request","arguments":{"action":"compare","group":"sum","size":64}}}]},"done":true}`)
	}))
	defer model.Close()
	client, err := newLocalModelClient(model.URL, "test-model", time.Second, false)
	require.NoError(t, err)
	app.model = client

	response := postUI(app, "/api/ask", `{"question":"Which two strategies could we put head to head for a total?"}`)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var preview uiPreview
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	require.True(t, preview.Ollama)
	require.Equal(t, 64, preview.Size)
	entries, err := os.ReadDir(app.root)
	require.NoError(t, err)
	require.Empty(t, entries)
}

type cancelingUIWorker struct{ started chan struct{} }

func (worker cancelingUIWorker) Execute(ctx context.Context, _ labworker.Request) (labworker.RunResult, error) {
	close(worker.started)
	<-ctx.Done()
	return labworker.RunResult{}, ctx.Err()
}

func TestLocalUIRejectsSimultaneousRunAndCancelsOnDisconnect(t *testing.T) {
	app := testLocalUI(t)
	preview := getUIStarter(t, app)
	worker := cancelingUIWorker{started: make(chan struct{})}
	tools, err := labtools.New(labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy(), worker)
	require.NoError(t, err)
	plan, err := tools.ProposeExperiment(additionStarter())
	require.NoError(t, err)
	app.pending.plan, app.pending.tools = plan, tools

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := uiRequest(app, "/api/run", approvalJSON(t, preview.Ref)).WithContext(ctx)
	first := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		app.ServeHTTP(first, req)
	}()
	select {
	case <-worker.started:
	case <-ctx.Done():
		t.Fatal("worker did not start")
	}
	require.Equal(t, http.StatusConflict, postUI(app, "/api/run", approvalJSON(t, preview.Ref)).Code)
	cancel()
	<-finished
	require.Contains(t, first.Body.String(), "measurement interrupted")
	require.NotEqual(t, preview.Ref, getUIStarter(t, app).Ref)
}

func testLocalUI(t *testing.T) *localUI {
	t.Helper()
	service, err := labservice.NewDeterministic(platform.Current(), labapi.DefaultPlanPolicy())
	require.NoError(t, err)
	app, err := newLocalUI("127.0.0.1:1234", t.TempDir(), service, nil)
	require.NoError(t, err)
	return app
}

func postUI(app *localUI, path, data string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, uiRequest(app, path, data))
	return recorder
}

func uiRequest(app *localUI, path, data string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, app.origin+path, strings.NewReader(data))
	request.Header.Set("X-Simdlab-Session", app.token)
	request.Header.Set("Origin", app.origin)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func getUIStarter(t *testing.T, app *localUI) uiPreview {
	t.Helper()
	response := postUI(app, "/api/starter", "")
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var preview uiPreview
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	return preview
}

func approvalJSON(t *testing.T, ref string) string {
	t.Helper()
	data, err := json.Marshal(map[string]string{"ref": ref})
	require.NoError(t, err)
	return string(data)
}
