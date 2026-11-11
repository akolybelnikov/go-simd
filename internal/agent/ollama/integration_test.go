package ollama_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/agent/ollama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupModelClient(t *testing.T) *ollama.Client {
	t.Helper()

	if os.Getenv("GO_SIMD_OLLAMA_INTEGRATION") != "1" {
		t.Skip("env var GO_SIMD_OLLAMA_INTEGRATION is not set")
	}

	ep, err := url.Parse("http://127.0.0.1:11434")
	require.NoError(t, err)

	httpClient := &http.Client{Timeout: 2 * time.Minute}
	modelClient, err := ollama.NewClient(ep, "qwen3:4b", httpClient)
	require.NoError(t, err)

	return modelClient
}

func TestIntegration_RuntimeText(t *testing.T) {
	modelClient := setupModelClient(t)
	agentRuntime, err := agent.NewRuntime(modelClient, "Answer briefly and directly", nil)
	require.NoError(t, err)
	require.NotNil(t, agentRuntime)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	agentMsg, err := agentRuntime.RunModelStep(ctx, []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "How do programs greet the world?",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, agent.RoleAssistant, agentMsg.Role)
	assert.NotEmpty(t, agentMsg.Content)

	t.Logf("LLM reply: %s", agentMsg.Content)
}

func TestIntegration_RuntimeToolCall(t *testing.T) {
	modelClient := setupModelClient(t)
	toolDefinitions := []agent.ToolDefinition{
		{
			Name:        "search_lab_docs",
			Description: "search the wiki and the catalogue for the requested operation",
			InputSchema: []byte(`{"type": "object", "properties": {"op_name": {"type": "string"}},"required": ["op_name"],"additionalProperties":false}`),
		},
	}

	agentRuntime, err := agent.NewRuntime(modelClient, "use search_lab_docs for operation-documentation questions instead of answering from memory.", toolDefinitions)
	require.NoError(t, err)
	require.NotNil(t, agentRuntime)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	agentMsg, err := agentRuntime.RunModelStep(ctx, []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "Use the documentation search tool to find information about addition.",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, agent.RoleAssistant, agentMsg.Role)
	require.Len(t, agentMsg.ToolCalls, 1)
	assert.Equal(t, "search_lab_docs", agentMsg.ToolCalls[0].Name)
	assert.JSONEq(t, `{"op_name": "addition"}`, string(agentMsg.ToolCalls[0].Arguments))
	t.Logf("model returned tool ToolCall: %+v", agentMsg.ToolCalls[0])
}
