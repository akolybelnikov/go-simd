package agent_test

import (
	"context"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRuntime(t *testing.T) {
	tests := []struct {
		name         string
		client       agent.ModelClient
		instructions string
		tools        []agent.ToolDefinition
		wantErr      bool
	}{
		{
			name:         "client and instructions with no tools",
			client:       &fakeModelClient{},
			instructions: "test instructions",
		},
		{
			name:         "with one valid tool",
			client:       &fakeModelClient{},
			instructions: "test instructions",
			tools: []agent.ToolDefinition{
				{
					Name:        "test_tool",
					Description: "A test tool",
					InputSchema: []byte(`{"type": "object"}`),
				},
			},
		},
		{
			name:         "nil client",
			instructions: "test instructions",
			wantErr:      true,
		},
		{
			name:    "empty instructions",
			client:  &fakeModelClient{},
			wantErr: true,
		},
		{
			name:         "invalid tool definition with no name",
			client:       &fakeModelClient{},
			instructions: "test instructions",
			tools: []agent.ToolDefinition{
				{
					Description: "A test tool",
					InputSchema: []byte(`{"type": "object"}`),
				},
			},
			wantErr: true,
		},
		{
			name:         "invalid tool definitions with duplicate tool names",
			client:       &fakeModelClient{},
			instructions: "test instructions",
			tools: []agent.ToolDefinition{
				{
					Name:        "duplicate_tool",
					Description: "First tool",
					InputSchema: []byte(`{"type": "object"}`),
				},
				{
					Name:        "duplicate_tool",
					Description: "Second tool",
					InputSchema: []byte(`{"type": "object"}`),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := agent.NewRuntime(tt.client, tt.instructions, tt.tools)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewRuntime() failed: %v", gotErr)
				}
				if got != nil {
					t.Fatal("Unexpected Runtime with error in return")
				}
				return
			}

			if tt.wantErr {
				t.Fatal("NewRuntime() succeeded unexpectedly")
			}

			if got == nil {
				t.Fatal("NewRuntime() returned nil Runtime unexpectedly")
			}
		})
	}
}

func TestRuntime_RunModelStep_Text(t *testing.T) {
	client := &fakeModelClient{
		events: []agent.ModelEvent{
			{
				Kind: agent.EventTextDelta,
				Text: "Hel",
			},
			{
				Kind: agent.EventTextDelta,
				Text: "lo",
			},
			{
				Kind: agent.EventCompleted,
			},
		},
	}

	runTime, err := agent.NewRuntime(client, "Greet the world", []agent.ToolDefinition{})
	require.NoError(t, err)

	msg, err := runTime.RunModelStep(context.Background(), []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "What's the greeting?",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, agent.RoleAssistant, msg.Role)
	assert.Equal(t, "Hello", msg.Content)
	assert.Nil(t, msg.ToolCalls)
}

func TestRuntime_RunModelStep_RequiresCompletion(t *testing.T) {
	client := &fakeModelClient{
		events: []agent.ModelEvent{
			{
				Kind: agent.EventTextDelta,
				Text: "Hel",
			},
			{
				Kind: agent.EventTextDelta,
				Text: "lo, ",
			},
			{
				Kind: agent.EventTextDelta,
				Text: "World!",
			},
		},
	}

	runTime, err := agent.NewRuntime(client, "Greet the world", []agent.ToolDefinition{})
	require.NoError(t, err)

	msg, err := runTime.RunModelStep(context.Background(), []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "How do programms greet the world?",
		},
	})
	assert.Error(t, err)
	assert.Zero(t, msg)
}

func TestRuntime_RunModelStep_RejectsEventAfterCompletion(t *testing.T) {
	client := &fakeModelClient{
		events: []agent.ModelEvent{
			{
				Kind: agent.EventTextDelta,
				Text: "Hello",
			},
			{
				Kind: agent.EventCompleted,
			},
			{
				Kind: agent.EventTextDelta,
				Text: "some unexpected text",
			},
		},
	}

	runTime, err := agent.NewRuntime(client, "Greet the world", nil)
	require.NoError(t, err)

	msg, err := runTime.RunModelStep(context.Background(), []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "What's the greeting?",
		},
	})
	assert.Error(t, err)
	assert.Zero(t, msg)
}

func TestRuntime_RunModelStep_ToolCall(t *testing.T) {
	client := &fakeModelClient{
		events: []agent.ModelEvent{
			{
				Kind: agent.EventToolCall,
				ToolCall: &agent.ToolCall{
					ID:        "12345",
					Name:      "search_lab_docs",
					Arguments: []byte(`{"op_name": "addition"}`),
				},
			},
			{
				Kind: agent.EventCompleted,
			},
		},
	}

	runTime, err := agent.NewRuntime(client, "search for the operation and explain the experiment", []agent.ToolDefinition{
		{
			Name:        "search_lab_docs",
			Description: "Search lab wiki for the requested term",
			InputSchema: []byte(`{"type": "object", "properties": {"op_name": {"type": "string"}}}`),
		},
	})
	require.NoError(t, err)

	msg, err := runTime.RunModelStep(context.Background(), []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "How to benchmark addition?",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, agent.RoleAssistant, msg.Role)
	assert.Empty(t, msg.Content)
	require.Len(t, msg.ToolCalls, 1)
	assert.Equal(t, agent.ToolCall{
		ID:        "12345",
		Name:      "search_lab_docs",
		Arguments: []byte(`{"op_name": "addition"}`),
	}, msg.ToolCalls[0])
}

func TestRuntime_RunModelStep_RejectsEmptyResponse(t *testing.T) {
	client := &fakeModelClient{
		events: []agent.ModelEvent{
			{
				Kind: agent.EventCompleted,
			},
		},
	}

	runTime, err := agent.NewRuntime(client, "search for the operation and explain the experiment", []agent.ToolDefinition{
		{
			Name:        "search_lab_docs",
			Description: "Search lab wiki for the requested term",
			InputSchema: []byte(`{"type": "object", "properties": {"op_name": {"type": "string"}}}`),
		},
	})
	require.NoError(t, err)

	msg, err := runTime.RunModelStep(context.Background(), []agent.Message{
		{
			Role:    agent.RoleUser,
			Content: "How to benchmark addition?",
		},
	})
	require.Error(t, err)
	assert.Zero(t, msg)
}
