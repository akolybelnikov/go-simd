package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

func TestMessageRole_Valid(t *testing.T) {
	tests := []struct {
		name string
		role agent.MessageRole
		want bool
	}{
		{
			name: "valid user role",
			role: agent.RoleUser,
			want: true,
		},
		{
			name: "valid assistant role",
			role: agent.RoleAssistant,
			want: true,
		},
		{
			name: "valid tool role",
			role: agent.RoleTool,
			want: true,
		},
		{
			name: "invalid role",
			role: "invalid",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.Valid()
			if got != tt.want {
				t.Fatalf("MessageRole.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelEventKind_Valid(t *testing.T) {
	tests := []struct {
		name  string
		event agent.ModelEventKind
		want  bool
	}{
		{
			name:  "valid text delta event",
			event: agent.EventTextDelta,
			want:  true,
		},
		{
			name:  "valid tool call event",
			event: agent.EventToolCall,
			want:  true,
		},
		{
			name:  "valid completed event",
			event: agent.EventCompleted,
			want:  true,
		},
		{
			name:  "invalid event",
			event: "invalid_kind",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.Valid()
			if got != tt.want {
				t.Fatalf("ModelEventKind.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   agent.ModelEvent
		wantErr bool
	}{
		{
			name: "valid text delta event",
			event: agent.ModelEvent{
				Kind: agent.EventTextDelta,
				Text: "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "valid tool call event",
			event: agent.ModelEvent{
				Kind: agent.EventToolCall,
				ToolCall: &agent.ToolCall{
					ID:        "123",
					Name:      "example_tool",
					Arguments: []byte(`{"arg1": "value1"}`),
				},
			},
			wantErr: false,
		},
		{
			name: "valid completed event",
			event: agent.ModelEvent{
				Kind: agent.EventCompleted,
			},
			wantErr: false,
		},
		{
			name: "unsupported event kind",
			event: agent.ModelEvent{
				Kind: "unsupported_kind",
			},
			wantErr: true,
		},
		{
			name: "invalid text delta event with missing text",
			event: agent.ModelEvent{
				Kind: agent.EventTextDelta,
			},
			wantErr: true,
		},
		{
			name: "invalid text delta event with unexpected tool call",
			event: agent.ModelEvent{
				Kind: agent.EventTextDelta,
				Text: "Hello, world!",
				ToolCall: &agent.ToolCall{
					ID:        "123",
					Name:      "example_tool",
					Arguments: []byte(`{"arg1": "value1"}`),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool call event with unexpected text",
			event: agent.ModelEvent{
				Kind: agent.EventToolCall,
				Text: "Unexpected text",
				ToolCall: &agent.ToolCall{
					ID:        "123",
					Name:      "example_tool",
					Arguments: []byte(`{"arg1": "value1"}`),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool call event with missing tool call",
			event: agent.ModelEvent{
				Kind: agent.EventToolCall,
			},
			wantErr: true,
		},
		{
			name: "invalid completed event with unexpected text",
			event: agent.ModelEvent{
				Kind: agent.EventCompleted,
				Text: "Unexpected text",
			},
			wantErr: true,
		},
		{
			name: "invalid completed event with unexpected tool call",
			event: agent.ModelEvent{
				Kind: agent.EventCompleted,
				ToolCall: &agent.ToolCall{
					ID:        "123",
					Name:      "example_tool",
					Arguments: []byte(`{"arg1": "value1"}`),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool call event with structurally invalid tool call",
			event: agent.ModelEvent{
				Kind: agent.EventToolCall,
				ToolCall: &agent.ToolCall{
					ID:   "123",
					Name: "example_tool",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.event.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestToolCall_Validate(t *testing.T) {
	tests := []struct {
		name    string
		call    agent.ToolCall
		wantErr bool
	}{
		{
			name:    "valid empty argument object",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`{}`)},
			wantErr: false,
		},
		{
			name:    "valid populated argument object",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`{"arg1": "value1"}`)},
			wantErr: false,
		},
		{
			name:    "missing ID",
			call:    agent.ToolCall{Name: "example_tool", Arguments: []byte(`{}`)},
			wantErr: true,
		},
		{
			name:    "missing name",
			call:    agent.ToolCall{ID: "123", Arguments: []byte(`{}`)},
			wantErr: true,
		},
		{
			name:    "missing arguments",
			call:    agent.ToolCall{ID: "123", Name: "example_tool"},
			wantErr: true,
		},
		{
			name:    "malformed JSON",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`{"arg1":}`)},
			wantErr: true,
		},
		{
			name:    "null arguments",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`null`)},
			wantErr: true,
		},
		{
			name:    "JSON array arguments",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`["value1"]`)},
			wantErr: true,
		},
		{
			name:    "JSON scalar arguments",
			call:    agent.ToolCall{ID: "123", Name: "example_tool", Arguments: []byte(`"value1"`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.call.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

type fakeModelClient struct {
	events   []agent.ModelEvent
	requests []agent.ModelRequest
}

var _ agent.ModelClient = (*fakeModelClient)(nil)

func (c *fakeModelClient) Stream(ctx context.Context, req agent.ModelRequest) (agent.ModelStream, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	c.requests = append(c.requests, req)

	return &fakeModelStream{events: c.events}, nil
}

func TestMessage_Validate(t *testing.T) {
	toolCall := agent.ToolCall{
		ID:        "123",
		Name:      "example_tool",
		Arguments: []byte(`{"arg1": "value1"}`),
	}
	tests := []struct {
		name    string
		message agent.Message
		wantErr bool
	}{
		{
			name: "user with content",
			message: agent.Message{
				Role:    agent.RoleUser,
				Content: "Hello, world!",
			},
		},
		{
			name: "assistant with text",
			message: agent.Message{
				Role:    agent.RoleAssistant,
				Content: "Hello, world!",
			},
		},
		{
			name: "assistant with a valid tool call",
			message: agent.Message{
				Role:      agent.RoleAssistant,
				ToolCalls: []agent.ToolCall{toolCall},
			},
		},
		{
			name: "assistant with both text and a valid tool call",
			message: agent.Message{
				Role:      agent.RoleAssistant,
				Content:   "Hello, world!",
				ToolCalls: []agent.ToolCall{toolCall},
			},
		},
		{
			name: "tool result with content and matching tool call ID",
			message: agent.Message{
				Role:       agent.RoleTool,
				Content:    "Tool result",
				ToolCallID: "123",
			},
		},
		{
			name: "unsupported role",
			message: agent.Message{
				Role:    "unsupported_role",
				Content: "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "user with missing content",
			message: agent.Message{
				Role: agent.RoleUser,
			},
			wantErr: true,
		},
		{
			name: "user containing tool calls",
			message: agent.Message{
				Role:      agent.RoleUser,
				Content:   "Hello, world!",
				ToolCalls: []agent.ToolCall{toolCall},
			},
			wantErr: true,
		},
		{
			name: "user containing a tool call ID",
			message: agent.Message{
				Role:       agent.RoleUser,
				Content:    "Hello, world!",
				ToolCallID: "123",
			},
			wantErr: true,
		},
		{
			name: "assistant with missing content and no tool calls",
			message: agent.Message{
				Role: agent.RoleAssistant,
			},
			wantErr: true,
		},
		{
			name: "assistant containing a tool call ID",
			message: agent.Message{
				Role:       agent.RoleAssistant,
				Content:    "Hello, world!",
				ToolCallID: "123",
			},
			wantErr: true,
		},
		{
			name: "assistant with an invalid tool call",
			message: agent.Message{
				Role:      agent.RoleAssistant,
				ToolCalls: []agent.ToolCall{{ID: "123", Name: "example_tool"}},
			},
			wantErr: true,
		},
		{
			name: "tool result with missing content",
			message: agent.Message{
				Role:       agent.RoleTool,
				ToolCallID: "123",
			},
			wantErr: true,
		},
		{
			name: "tool result with missing tool call ID",
			message: agent.Message{
				Role:    agent.RoleTool,
				Content: "Tool result",
			},
			wantErr: true,
		},
		{
			name: "tool result containing tool calls",
			message: agent.Message{
				Role:       agent.RoleTool,
				Content:    "Tool result",
				ToolCallID: "123",
				ToolCalls:  []agent.ToolCall{toolCall},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.message.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestToolDefinition_Validate(t *testing.T) {
	tests := []struct {
		name       string
		definition agent.ToolDefinition
		wantErr    bool
	}{
		{
			name: "object schema with no properties",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"type": "object", "properties": {}}`),
			},
			wantErr: false,
		},
		{
			name: "object schema with properties",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"type": "object", "properties": {"param1": {"type": "string"}}}`),
			},
			wantErr: false,
		},
		{
			name: "missing tool name",
			definition: agent.ToolDefinition{
				Description: "An example tool",
				InputSchema: []byte(`{"type": "object", "properties": {}}`),
			},
			wantErr: true,
		},
		{
			name: "missing tool description",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				InputSchema: []byte(`{"type": "object", "properties": {}}`),
			},
			wantErr: true,
		},
		{
			name: "missing input schema",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
			},
			wantErr: true,
		},
		{
			name: "invalid input schema",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"type": "object", "properties": {`), // malformed JSON
			},
			wantErr: true,
		},
		{
			name: "JSON null input schema",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`null`), // JSON null
			},
			wantErr: true,
		},
		{
			name: "input schema root is an array",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`[]`), // JSON array
			},
			wantErr: true,
		},
		{
			name: "input schema missing 'type' field",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"properties": {}}`), // missing 'type' field
			},
			wantErr: true,
		},
		{
			name: "input schema with array as 'type' field",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"type": "array", "properties": {}}`), // 'type' field is `array` instead of `object`
			},
			wantErr: true,
		},
		{
			name: "input schema with non-string 'type' field",
			definition: agent.ToolDefinition{
				Name:        "example_tool",
				Description: "An example tool",
				InputSchema: []byte(`{"type": 123, "properties": {}}`), // 'type' field is not a string
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.definition.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestModelRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     agent.ModelRequest
		wantErr bool
	}{
		{
			name: "valid request with instructions and one message",
			req: agent.ModelRequest{
				Instructions: "Please summarize the following conversation.",
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Hello, how are you?",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with instructions, valid history and a valid tool",
			req: agent.ModelRequest{
				Instructions: "Explain the results of the experiment",
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Execute the experiment and provide the results.",
					},
					{
						Role: agent.RoleAssistant,
						ToolCalls: []agent.ToolCall{
							{
								ID:        "12345",
								Name:      "example_tool",
								Arguments: []byte(`{"name": "test_XI"}`),
							},
						},
					},
					{
						Role:       agent.RoleTool,
						Content:    "Tool result",
						ToolCallID: "12345",
					},
				},
				Tools: []agent.ToolDefinition{
					{
						Name:        "example_tool",
						Description: "An example tool",
						InputSchema: []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing instructions",
			req: agent.ModelRequest{
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Hello, how are you?",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing messages",
			req: agent.ModelRequest{
				Instructions: "Please summarize the following conversation.",
			},
			wantErr: true,
		},
		{
			name: "invalid message in history",
			req: agent.ModelRequest{
				Instructions: "Please summarize the following conversation.",
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Hello, how are you?",
					},
					{
						Role:       agent.RoleTool,
						ToolCallID: "12345",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid tool definition",
			req: agent.ModelRequest{
				Instructions: "Please summarize the following conversation.",
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Hello, how are you?",
					},
				},
				Tools: []agent.ToolDefinition{
					{
						Name:        "example_tool",
						Description: "An example tool",
						InputSchema: []byte(`{"type": "array"}`), // invalid schema
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate valid tool names",
			req: agent.ModelRequest{
				Instructions: "Please summarize the following conversation.",
				Messages: []agent.Message{
					{
						Role:    agent.RoleUser,
						Content: "Hello, how are you?",
					},
				},
				Tools: []agent.ToolDefinition{
					{
						Name:        "example_tool",
						Description: "An example tool",
						InputSchema: []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`),
					},
					{
						Name:        "example_tool", // duplicate name
						Description: "Another example tool",
						InputSchema: []byte(`{"type": "object", "properties": {"name": {"type": "string"}}}`),
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.req.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

type fakeModelStream struct {
	events []agent.ModelEvent
	index  int
	closed bool
}

var _ agent.ModelStream = (*fakeModelStream)(nil)

func (f *fakeModelStream) Next() (agent.ModelEvent, error) {
	if f.index >= len(f.events) {
		return agent.ModelEvent{}, io.EOF
	}

	event := f.events[f.index]
	f.index++

	return event, nil
}

func (f *fakeModelStream) Close() error {
	f.closed = true

	return nil
}
