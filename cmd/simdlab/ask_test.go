package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/labapi"
)

func TestAskReplaysRealCatalog(t *testing.T) {
	requests := make(chan []byte, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/chat" {
			http.NotFound(w, r)
			return
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		select {
		case requests <- data:
		default:
			http.Error(w, "too many requests", http.StatusBadRequest)
			return
		}
		if len(requests) == 1 {
			_, _ = io.WriteString(w, `{"message":{"tool_calls":[{"function":{"name":"explore_catalog","arguments":{"group":"addition"}}}]},"done":true}`)
			return
		}
		_, _ = io.WriteString(w, `{"message":{"content":"Scalar is available."},"done":true}`)
	}))
	defer server.Close()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"ask", "-preview", "-endpoint", server.URL, "-model", "fixture", "Which addition strategies exist?"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("ask code=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "Scalar is available.\n" {
		t.Fatalf("answer = %q", got)
	}
	if len(requests) != 2 {
		t.Fatalf("request count = %d", len(requests))
	}
	<-requests
	var request struct {
		Model    string `json:"model"`
		Messages []struct {
			Role     string `json:"role"`
			ToolName string `json:"tool_name"`
			Content  string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(<-requests, &request); err != nil || request.Model != "fixture" || len(request.Messages) != 4 {
		t.Fatalf("unexpected follow-up request: %+v, %v", request, err)
	}
	result := request.Messages[3]
	var catalog labapi.CatalogResponse
	if err := json.Unmarshal([]byte(result.Content), &catalog); err != nil || result.Role != "tool" || result.ToolName != "explore_catalog" || len(catalog.Operations) != 1 || catalog.Operations[0].Group != "addition" {
		t.Fatalf("unexpected catalog result: %+v, %v", result, err)
	}
}
