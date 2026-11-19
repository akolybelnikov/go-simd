package ollama

import (
	"io"
	"strings"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ agent.ModelStream = (*modelStream)(nil)

func TestStreamErrorsAndLimits(t *testing.T) {
	for _, tt := range []struct{ name, data, want string }{
		{"provider error", `{"error":"provider diagnostic"}`, "reported an error"},
		{"malformed JSON", `{"message":`, "decoding body"},
		{"oversized response", strings.Repeat(" ", maxResponseBytes+1), "response exceeds"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			stream := newModelStream(io.NopCloser(strings.NewReader(tt.data)))
			_, err := stream.Next()
			require.ErrorContains(t, err, tt.want)
			require.NoError(t, stream.Close())
		})
	}
}

func TestStreamAssignsDistinctIDsAcrossChunks(t *testing.T) {
	chunk := `{"message":{"tool_calls":[{"function":{"index":0,"name":"explore_catalog","arguments":{"group":"all"}}}]}}`
	stream := newModelStream(io.NopCloser(strings.NewReader(chunk + "\n" + chunk)))
	for _, id := range []string{"ollama-call-0", "ollama-call-1"} {
		event, err := stream.Next()
		require.NoError(t, err)
		require.NotNil(t, event.ToolCall)
		require.Equal(t, id, event.ToolCall.ID)
	}
	require.NoError(t, stream.Close())
}

func Test_Streaming(t *testing.T) {
	streamData := `{"message":{"role":"assistant","content":"","thinking":"private reasoning"},"done":false}
	{"message":{"role":"assistant","content":"Hel"},"done":false}
	{"message":{"role":"assistant","content":"lo"},"done":false}
	{"message":{"role":"assistant","content":""},"done":true}`

	body := io.NopCloser(strings.NewReader(streamData))
	stream := newModelStream(body)

	event, err := stream.Next()
	require.NoError(t, err)
	assert.Equal(t, agent.EventTextDelta, event.Kind)
	assert.Equal(t, "Hel", event.Text)

	event, err = stream.Next()
	require.NoError(t, err)
	assert.Equal(t, agent.EventTextDelta, event.Kind)
	assert.Equal(t, "lo", event.Text)

	event, err = stream.Next()
	require.NoError(t, err)
	assert.Equal(t, agent.EventCompleted, event.Kind)
	assert.Empty(t, event.Text)

	event, err = stream.Next()
	require.ErrorIs(t, err, io.EOF)
	assert.Zero(t, event)

	require.NoError(t, stream.Close())
}

func Test_StreamingToolCall(t *testing.T) {
	streamData := `{"message":{"role":"assistant","content":"","tool_calls":[{"type":"function","function":{"index":0,"name":"search_lab_docs","arguments":{"op_name": "addition"}}}]},"done":true}`

	body := io.NopCloser(strings.NewReader(streamData))
	stream := newModelStream(body)

	event, err := stream.Next()
	require.NoError(t, err)
	assert.Equal(t, agent.EventToolCall, event.Kind)
	require.NotNil(t, event.ToolCall)
	assert.Equal(t, "ollama-call-0", event.ToolCall.ID)
	assert.Equal(t, "search_lab_docs", event.ToolCall.Name)
	assert.JSONEq(t, `{"op_name": "addition"}`, string(event.ToolCall.Arguments))

	event, err = stream.Next()
	require.NoError(t, err)
	assert.Equal(t, agent.EventCompleted, event.Kind)
	assert.Empty(t, event.Text)

	event, err = stream.Next()
	require.ErrorIs(t, err, io.EOF)
	assert.Zero(t, event)

	require.NoError(t, stream.Close())
}
