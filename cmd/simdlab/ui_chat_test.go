package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalUIChatUsesSelectedFactsAndCannotApprove(t *testing.T) {
	app := testLocalUI(t)
	starter := getUIStarter(t, app)
	data, err := os.ReadFile(filepath.Join("testdata", "sum-cancellation.measurements.json"))
	require.NoError(t, err)
	opened := postUI(app, "/api/open", string(data))
	require.Equal(t, http.StatusOK, opened.Code)
	var result uiResult
	require.NoError(t, json.Unmarshal(opened.Body.Bytes(), &result))
	require.NotEmpty(t, result.ChatRef)

	request := uiChatRequest{
		Provider: "openai", Selection: chatSelection{ResultRef: result.ChatRef, Size: result.Series[0].Size},
		Question: "What was measured?",
		History:  []chatMessage{{"user", "What is SIMD?"}, {"assistant", "It works on several values together."}},
	}
	payload, err := json.Marshal(request)
	require.NoError(t, err)
	contextPreview := postUI(app, "/api/chat/context", string(payload))
	require.Equal(t, http.StatusOK, contextPreview.Code, contextPreview.Body.String())
	var shared chatContext
	require.NoError(t, json.Unmarshal(contextPreview.Body.Bytes(), &shared))
	require.NotEmpty(t, shared.Result.Facts)
	fact := shared.Result.Facts[0]

	requests := make(chan *http.Request, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		var sent struct {
			Model           string `json:"model"`
			Store           *bool  `json:"store"`
			Reasoning       map[string]string
			Input           []chatMessage
			MaxOutputTokens int `json:"max_output_tokens"`
			Tools           json.RawMessage
		}
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Error(err)
			return
		}
		if sent.Model != openAIChatModel || sent.Store == nil || *sent.Store ||
			sent.Reasoning["effort"] != "low" || sent.MaxOutputTokens <= 0 || sent.Tools != nil {
			t.Errorf("unexpected OpenAI request settings: %+v", sent)
		}
		if len(sent.Input) != 3 || sent.Input[0] != request.History[0] || sent.Input[1] != request.History[1] {
			t.Errorf("unexpected history: %+v", sent.Input)
		}
		if len(sent.Input) == 3 {
			prompt := sent.Input[2].Content
			for _, private := range []string{app.root, app.token, result.ChatRef, "test-secret", "nanoseconds_per_op"} {
				if strings.Contains(prompt, private) {
					t.Errorf("unexpected private or raw field in model context: %q", private)
				}
			}
			if !strings.Contains(prompt, fact.ID) {
				t.Error("selected fact missing from model context")
			}
		}
		answer, err := json.Marshal(chatAnswer{
			Answer: "The measurements describe generated data.", Hypothesis: "Try a larger batch.",
			FactIDs: []string{fact.ID}, Draft: &chatDraft{"sum", "simd-striped", 80000, 30},
		})
		if err != nil {
			t.Error(err)
			return
		}
		writeOpenAIChatTestReply(w, string(answer))
	}))
	defer upstream.Close()
	provider := newChatProvider("openai", openAIChatModel, "test-secret")
	provider.endpoint = upstream.URL
	app.chatProviders = []chatProvider{provider}

	unauthenticated := uiRequest(app, "/api/chat", string(payload))
	unauthenticated.Header.Del("X-Simdlab-Session")
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, unauthenticated)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Empty(t, requests)

	response := postUI(app, "/api/chat", string(payload))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, "Bearer test-secret", (<-requests).Header.Get("Authorization"))
	var reply uiChatReply
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &reply))
	require.Equal(t, []chatFact{fact}, reply.Facts)
	require.Equal(t, &chatDraft{"sum", "simd-striped", 80000, 30}, reply.Draft)
	require.Equal(t, starter.Ref, app.pending.ref)
	entries, err := os.ReadDir(app.root)
	require.NoError(t, err)
	require.Empty(t, entries)

	require.Equal(t, http.StatusOK, postUI(app, "/api/open", string(data)).Code)
	require.Equal(t, http.StatusConflict, postUI(app, "/api/chat", string(payload)).Code)
	require.Empty(t, requests)
}

func TestLocalUIChatRejectsInvalidFactAndDraft(t *testing.T) {
	app := testLocalUI(t)
	result := &chatResultContext{Group: "addition", Facts: []chatFact{{"f1", "server-rendered"}}}
	for _, raw := range []string{
		`{"answer":"A","fact_ids":[],"hypothesis":"","draft":{"group":"addition","method":"scalar","size":8,"repeats":3}}`,
		`{"answer":"A","fact_ids":[],"hypothesis":"","draft":{"group":"addition","method":"simd-partial","size":16777217,"repeats":3}}`,
		`{"answer":"A","fact_ids":[],"hypothesis":"","draft":{"group":"addition","method":"simd-partial","size":8,"repeats":51}}`,
		`{"answer":"A","fact_ids":[],"hypothesis":"","draft":{"group":"sum","method":"simd-striped","size":8,"repeats":3}}`,
	} {
		reply, err := app.checkedChatAnswer(raw, result)
		require.NoError(t, err)
		require.Nil(t, reply.Draft)
		require.NotEmpty(t, reply.DraftError)
	}
	for _, raw := range []string{
		`{"answer":"A","fact_ids":["other-run"],"hypothesis":"","draft":null}`,
		`{"answer":"A","fact_ids":[],"hypothesis":"","draft":null,"ref":"approval"}`,
	} {
		_, err := app.checkedChatAnswer(raw, result)
		require.Error(t, err, raw)
	}
	require.Nil(t, app.pending)
	app.hostedTurns = chatSessionLimit
	status, _ := app.beginModel("openai")
	require.Equal(t, http.StatusTooManyRequests, status)
}

func TestLocalUIChatCancellationReleasesMeasurementGate(t *testing.T) {
	app := testLocalUI(t)
	starter := getUIStarter(t, app)
	inside := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			_, _ = w.Write([]byte(`{"models":[{"name":"example:latest"}]}`))
			return
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if string(body["think"]) != "false" || body["format"] == nil {
			t.Errorf("unexpected Ollama settings: %+v", body)
		}
		close(inside)
		<-r.Context().Done()
	}))
	defer upstream.Close()
	provider := newChatProvider("ollama", "example", "")
	provider.endpoint = upstream.URL
	app.chatProviders = []chatProvider{provider}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chat := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		request := uiRequest(app, "/api/chat", `{"provider":"ollama","question":"What is a scalar tail?"}`)
		app.ServeHTTP(chat, request.WithContext(ctx))
	}()
	select {
	case <-inside:
	case <-time.After(4 * time.Second):
		t.Fatal("Ollama request did not start")
	}
	require.Equal(t, http.StatusConflict, postUI(app, "/api/run", approvalJSON(t, starter.Ref)).Code)
	cancel()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("chat did not stop on disconnect")
	}
	require.Equal(t, http.StatusBadGateway, chat.Code)
	require.Equal(t, starter.Ref, app.pending.ref)
	require.False(t, app.chatRunning)
}

func TestOllamaChatRejectsInstalledRemoteModelBeforeSharing(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Error("question sent to remote model")
		}
		_, _ = w.Write([]byte(`{"models":[{"name":"remote:latest","remote_host":"https://ollama.example"}]}`))
	}))
	defer upstream.Close()
	provider := newChatProvider("ollama", "remote", "")
	provider.endpoint = upstream.URL
	_, err := provider.answer(t.Context(), nil, "Does this leave the machine?", nil)
	require.ErrorContains(t, err, "installed local model")
}

func TestOllamaChatKeepsAnswerAndServerEvidence(t *testing.T) {
	app := testLocalUI(t)
	capture, err := os.ReadFile(filepath.Join("testdata", "sum-cancellation.measurements.json"))
	require.NoError(t, err)
	opened := postUI(app, "/api/open", string(capture))
	require.Equal(t, http.StatusOK, opened.Code)
	var view uiResult
	require.NoError(t, json.Unmarshal(opened.Body.Bytes(), &view))
	selection := chatSelection{ResultRef: view.ChatRef, Size: view.Series[0].Size}
	lab, err := app.buildChatContext(selection, "")
	require.NoError(t, err)

	replies := []string{
		"```json\n" + `{"answer":"Answer about the selected run.","fact_ids":["invented-fact"],"hypothesis":["bad field"],"draft":{"group":"sum","method":"simd-striped","size":"80000","repeats":30}}` + "\n```",
		`{"answer":"Try a larger batch.","draft":{"group":"sum","method":"simd-striped","size":80000,"repeats":30}}`,
		`{"answer":`,
		`{"answer":`,
	}
	requests := make(chan string, len(replies))
	for _, reply := range replies {
		requests <- reply
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			_, _ = w.Write([]byte(`{"models":[{"name":"example:latest"}]}`))
			return
		}
		var payload struct {
			Format json.RawMessage `json:"format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || strings.Contains(string(payload.Format), "fact_ids") {
			t.Errorf("unexpected Ollama schema: %s (%v)", payload.Format, err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"done": true, "message": chatMessage{Content: <-requests}})
	}))
	defer upstream.Close()
	provider := newChatProvider("ollama", "example", "")
	provider.endpoint = upstream.URL
	app.chatProviders = []chatProvider{provider}

	request, err := json.Marshal(uiChatRequest{Provider: "ollama", Selection: selection, Question: "What can I conclude?"})
	require.NoError(t, err)
	first := postUI(app, "/api/chat", string(request))
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	var reply uiChatReply
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &reply))
	require.Equal(t, "Answer about the selected run.", reply.Answer)
	require.Empty(t, reply.Hypothesis)
	require.Equal(t, lab.Result.Facts, reply.Facts)
	require.NotContains(t, first.Body.String(), "invented-fact")
	require.Nil(t, reply.Draft)
	require.NotEmpty(t, reply.DraftError)

	valid := postUI(app, "/api/chat", string(request))
	require.Equal(t, http.StatusOK, valid.Code, valid.Body.String())
	reply = uiChatReply{}
	require.NoError(t, json.Unmarshal(valid.Body.Bytes(), &reply))
	require.Equal(t, &chatDraft{"sum", "simd-striped", 80000, 30}, reply.Draft)
	require.Nil(t, app.pending)

	fallback := postUI(app, "/api/chat", string(request))
	require.Equal(t, http.StatusOK, fallback.Code, fallback.Body.String())
	reply = uiChatReply{}
	require.NoError(t, json.Unmarshal(fallback.Body.Bytes(), &reply))
	require.True(t, reply.LabOnly)
	require.Nil(t, reply.Draft)
	require.Equal(t, lab.Result.Facts, reply.Facts)

	explore := postUI(app, "/api/chat", `{"provider":"ollama","question":"What is a scalar tail?"}`)
	require.Equal(t, http.StatusBadGateway, explore.Code)
	require.Contains(t, explore.Body.String(), "Copy to another assistant")
}

func writeOpenAIChatTestReply(w http.ResponseWriter, answer string) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "completed",
		"output": []any{map[string]any{"type": "message", "content": []any{
			map[string]string{"type": "output_text", "text": answer},
		}}},
	})
}
