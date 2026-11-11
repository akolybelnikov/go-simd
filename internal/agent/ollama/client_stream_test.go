package ollama_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/agent/ollama"
	"github.com/stretchr/testify/require"
)

func TestClient_StreamText(t *testing.T) {
	streamData := `{"message":{"role":"assistant","content":"Hel"},"done":false}
	{"message":{"role":"assistant","content":"lo"},"done":false}
	{"message":{"role":"assistant","content":""},"done":true}`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, streamData)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	client, err := ollama.NewClient(baseURL, "qwen3:4b", ts.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	req := agent.ModelRequest{Instructions: "Answer briefly", Messages: []agent.Message{{
		Role:    agent.RoleUser,
		Content: "Hello",
	}}}
	err = req.Validate()
	require.NoError(t, err)

	stream, err := client.Stream(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, stream)

	err = stream.Close()
	require.NoError(t, err)
}

func TestClient_StreamRejectsHTTPError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	baseURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	client, err := ollama.NewClient(baseURL, "qwen3:4b", ts.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	req := agent.ModelRequest{Instructions: "Answer briefly", Messages: []agent.Message{{
		Role:    agent.RoleUser,
		Content: "Hello",
	}}}
	err = req.Validate()
	require.NoError(t, err)

	stream, err := client.Stream(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, stream)
}
