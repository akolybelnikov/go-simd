package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalModelDiscoveryAndPerRequestSelection(t *testing.T) {
	var requested []string
	installed := true
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			if !installed {
				_, _ = w.Write([]byte(`{"models":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"models":[{"name":"qwen3.5:9b"},{"name":"custom:latest"},{"name":"remote:latest","remote_host":"https://example.org"},{"name":"other:cloud"},{"name":"bad;command"}]}`))
			return
		}
		var body struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		requested = append(requested, body.Model)
		_, _ = w.Write([]byte(`{"done":true,"message":{"content":"{\"answer\":\"A grounded reply\",\"draft\":null}"}}`))
	}))
	defer upstream.Close()
	app := testLocalUI(t)
	provider := newChatProvider("ollama", "", "")
	provider.endpoint = upstream.URL + "/api/chat"
	app.localChat = &provider
	app.initialChatModel = "qwen3.5:9b"

	unauthorized := uiRequest(app, "/api/chat/models", "")
	unauthorized.Header.Del("X-Simdlab-Session")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, unauthorized)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	response := postUI(app, "/api/chat/models", "")
	require.Equal(t, http.StatusOK, response.Code)
	require.NotContains(t, response.Body.String(), "remote:latest")
	require.NotContains(t, response.Body.String(), "other:cloud")
	require.NotContains(t, response.Body.String(), "bad;command")
	require.Contains(t, response.Body.String(), "not a browser-chat recommendation")
	require.Contains(t, response.Body.String(), "Not evaluated")
	for _, name := range []string{"custom:latest", "qwen3.5:9b"} {
		body, err := json.Marshal(uiChatRequest{Provider: "ollama", Model: name, Question: "What is SIMD?"})
		require.NoError(t, err)
		reply := postUI(app, "/api/chat", string(body))
		require.Equal(t, http.StatusOK, reply.Code, reply.Body.String())
	}
	require.Equal(t, []string{"custom:latest", "qwen3.5:9b"}, requested)
	require.Empty(t, app.localChat.model, "selection must not mutate the server template")
	for _, name := range []string{"missing", "remote:latest", "other:cloud"} {
		body, err := json.Marshal(uiChatRequest{Provider: "ollama", Model: name, Question: "What is SIMD?"})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadGateway, postUI(app, "/api/chat", string(body)).Code)
	}
	require.Len(t, requested, 2, "unavailable/remote selections must not receive the question")
	installed = false
	response = postUI(app, "/api/chat/models", "")
	require.Contains(t, response.Body.String(), `"installed":false`)
	require.Nil(t, app.pending)
}

func TestConfiguredRemoteModelIsSentAndKeyIsNotListed(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer test-api-secret", r.Header.Get("Authorization"))
		var body struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "configured-model", body.Model)
		writeOpenAIChatTestReply(w, `{"answer":"A","hypothesis":"","fact_ids":[],"draft":null}`)
	}))
	defer upstream.Close()
	app := testLocalUI(t)
	provider := newChatProvider("openai", "configured-model", "test-api-secret")
	provider.endpoint = upstream.URL
	app.chatProviders = []chatProvider{provider}
	catalog := postUI(app, "/api/catalog", "")
	require.Contains(t, catalog.Body.String(), "configured-model")
	require.NotContains(t, catalog.Body.String(), "test-api-secret")
	require.Equal(t, http.StatusOK, postUI(app, "/api/chat", `{"provider":"openai","question":"What is SIMD?"}`).Code)
}

func TestNoModelSelectedKeepsDeterministicGuideOffline(t *testing.T) {
	var stdout, stderr bytes.Buffer
	requests := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++ }))
	defer upstream.Close()
	require.Zero(t, run([]string{"ask", "-endpoint", upstream.URL}, strings.NewReader("Some unusual request\nquit\n"), &stdout, &stderr))
	require.Contains(t, stdout.String(), "No local model selected")
	require.Zero(t, requests)
	for _, args := range [][]string{
		{"ask", "-preview", "What is SIMD?"},
		{"experiment", "-preview", "Compare addition"},
		{"interpret", "-preview", "testdata/sum-cancellation.measurements.json"},
	} {
		stdout.Reset()
		stderr.Reset()
		require.Equal(t, 2, run(args, strings.NewReader(""), &stdout, &stderr))
		require.Contains(t, stderr.String(), "-model MODEL")
	}
}

func TestAISetupPageAndDialogGuidanceAreAvailable(t *testing.T) {
	app := testLocalUI(t)
	for _, path := range []string{"/", "/wiki/34-AI-setup-and-model-recommendations"} {
		req := httptest.NewRequest(http.MethodGet, app.origin+path, nil)
		req.Host = app.host
		response := httptest.NewRecorder()
		app.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code, path)
		if path == "/" {
			require.Contains(t, response.Body.String(), "We do not yet recommend a local model for browser chat")
			require.Contains(t, response.Body.String(), "/wiki/34-AI-setup-and-model-recommendations")
		} else {
			require.Contains(t, response.Body.String(), "no browser-chat recommendation yet")
			require.Contains(t, response.Body.String(), "-openai-model")
			require.Contains(t, response.Body.String(), `href="/wiki/33-Chat-with-an-agent"`)
		}
	}
}
