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

func Test_Streaming(t *testing.T) {
	streamData := `{"message":{"role":"assistant","content":"Hel"},"done":false}
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
