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
		think:     true,
	}

	var mRequest = agent.ModelRequest{
		Instructions: "Answer briefly",
		Messages:     []agent.Message{{Role: agent.RoleUser, Content: "Hello"}},
	}

	req, err := client.buildChatRequest(mRequest)
	require.NoError(t, err)
	assert.Equal(t, client.modelName, req.Model)
	assert.True(t, req.Stream)
	assert.True(t, req.Think)

	require.Len(t, req.Messages, 2)
	assert.Equal(t, systemRole, req.Messages[0].Role)
	assert.Equal(t, mRequest.Instructions, req.Messages[0].Content)
	assert.Equal(t, userRole, req.Messages[1].Role)
	assert.Equal(t, "Hello", req.Messages[1].Content)
}

func Test_BuildChatRequest_Tool(t *testing.T) {
	var client = Client{
		modelName: "qwen3:4b",
	}

	var mRequest = agent.ModelRequest{
		Instructions: "Answer briefly",
		Messages:     []agent.Message{{Role: agent.RoleUser, Content: "Hello"}},
		Tools: []agent.ToolDefinition{
			{
				Name:        "search_lab_docs",
				Description: "look throught eh catalogue of operations",
				InputSchema: []byte(`{"type": "object", "properties": {"op_name": {"type": "string"}}}`),
			},
		},
	}

	req, err := client.buildChatRequest(mRequest)
	require.NoError(t, err)
	assert.Equal(t, client.modelName, req.Model)
	require.Len(t, req.Tools, 1)
	assert.Equal(t, "function", req.Tools[0].Type)
	assert.Equal(t, mRequest.Tools[0].Name, req.Tools[0].Function.Name)
	assert.Equal(t, mRequest.Tools[0].Description, req.Tools[0].Function.Description)
	assert.JSONEq(t, string(mRequest.Tools[0].InputSchema), string(req.Tools[0].Function.Parameters))

}

func Test_BuildChatRequest_AssistantToolCall(t *testing.T) {
	client := Client{modelName: "qwen3:4b"}
	mRequest := agent.ModelRequest{
		Instructions: "Answer briefly",
		Messages: []agent.Message{
			{Role: agent.RoleUser, Content: "What can you tell me about addition?"},
			{
				Role: agent.RoleAssistant,
				ToolCalls: []agent.ToolCall{
					{
						ID:        "call-1",
						Name:      "explore_catalog",
						Arguments: []byte(`{"group":"addition"}`),
					},
				},
			},
		},
	}

	req, err := client.buildChatRequest(mRequest)
	require.NoError(t, err)
	require.Len(t, req.Messages, 3)
	assert.Equal(t, systemRole, req.Messages[0].Role)
	assert.Equal(t, userRole, req.Messages[1].Role)
	assert.Equal(t, assistantRole, req.Messages[2].Role)
	require.Len(t, req.Messages[2].ToolCalls, 1)
	assert.Equal(t, "explore_catalog", req.Messages[2].ToolCalls[0].Function.Name)
	assert.JSONEq(t, `{"group":"addition"}`, string(req.Messages[2].ToolCalls[0].Function.Arguments))
}

func Test_BuildChatRequest_ToolResult(t *testing.T) {
	tests := []struct {
		name              string
		messages          []agent.Message
		wantErr           bool
		wantRoles         []string
		wantToolName      string
		wantResultContent string
	}{
		{
			name: "matched result",
			messages: []agent.Message{
				{Role: agent.RoleUser, Content: "What can you tell me about addition?"},
				{
					Role: agent.RoleAssistant,
					ToolCalls: []agent.ToolCall{
						{
							ID:        "call-1",
							Name:      "explore_catalog",
							Arguments: []byte(`{"group":"addition"}`),
						},
					},
				},
				{
					Role:       agent.RoleTool,
					ToolCallID: "call-1",
					Content:    `{"result":"addition catalog"}`,
				},
			},
			wantRoles:         []string{systemRole, userRole, assistantRole, "tool"},
			wantToolName:      "explore_catalog",
			wantResultContent: `{"result":"addition catalog"}`,
		},
		{
			name: "unmatched result",
			messages: []agent.Message{
				{Role: agent.RoleUser, Content: "What can you tell me about addition?"},
				{
					Role:       agent.RoleTool,
					ToolCallID: "missing-call",
					Content:    `{"result":"addition catalog"}`,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{modelName: "qwen3:4b"}
			mRequest := agent.ModelRequest{
				Instructions: "Answer briefly",
				Messages:     tt.messages,
			}

			req, err := client.buildChatRequest(mRequest)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, req)
				return
			}

			require.NoError(t, err)
			require.Len(t, req.Messages, len(tt.wantRoles))
			for i, role := range tt.wantRoles {
				assert.Equal(t, role, req.Messages[i].Role)
			}
			assert.Equal(t, tt.wantToolName, req.Messages[3].ToolName)
			assert.Equal(t, tt.wantResultContent, req.Messages[3].Content)
		})
	}
}
