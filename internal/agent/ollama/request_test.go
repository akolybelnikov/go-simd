package ollama

import (
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BuildChatRequest(t *testing.T) {
	var client = Client{
		modelName: "qwen3:4b",
	}

	var mRequest = agent.ModelRequest{
		Instructions: "Answer briefly",
		Messages:     []agent.Message{{Role: agent.RoleUser, Content: "Hello"}},
	}

	req, err := client.buildChatRequest(mRequest)
	require.NoError(t, err)
	assert.Equal(t, client.modelName, req.Model)
	assert.True(t, req.Stream)
	assert.False(t, req.Think)

	require.Len(t, req.Messages, 2)
	assert.Equal(t, systemRole, req.Messages[0].Role)
	assert.Equal(t, mRequest.Instructions, req.Messages[0].Content)
	assert.Equal(t, userRole, req.Messages[1].Role)
	assert.Equal(t, "Hello", req.Messages[1].Content)
}
