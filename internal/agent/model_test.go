package agent_test

import (
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
