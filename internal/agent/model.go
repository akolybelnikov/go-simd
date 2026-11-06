// Package agent owns the provider-neutral conversation loop.
package agent

import (
	"encoding/json"
	"fmt"
)

// MessageRole is the role of a message in a conversation.
type MessageRole string

// These are supported message roles in a conversation.
const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// Valid reports if the role is supported.
func (r MessageRole) Valid() bool {
	switch r {
	case RoleUser, RoleAssistant, RoleTool:
		return true
	default:
		return false
	}
}

// ModelEventKind is the kind of event that can be emitted by the model.
type ModelEventKind string

// These are supported model event kinds.
const (
	EventTextDelta ModelEventKind = "text_delta"
	EventToolCall  ModelEventKind = "tool_call"
	EventCompleted ModelEventKind = "completed"
)

// Valid reports if the event kind is supported.
func (k ModelEventKind) Valid() bool {
	switch k {
	case EventTextDelta, EventToolCall, EventCompleted:
		return true
	default:
		return false
	}
}

// ToolCall is a request to call a tool with the given arguments.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// Validate returns an error if the tool call is structurally invalid.
func (t ToolCall) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("missing tool call ID")
	}
	if t.Name == "" {
		return fmt.Errorf("missing tool call name")
	}

	var args map[string]json.RawMessage
	if err := json.Unmarshal(t.Arguments, &args); err != nil {
		return fmt.Errorf("invalid tool call arguments: %w", err)
	}
	if args == nil {
		return fmt.Errorf("missing tool call arguments")
	}

	return nil
}

// Message is a single message in a conversation.
type Message struct {
	Role       MessageRole
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
}

// ToolDefinition is a description of a tool for an LLM call.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// ModelRequest is a request to the model with trusted instructions and the ordered conversation history.
type ModelRequest struct {
	Instructions string
	Messages     []Message
	Tools        []ToolDefinition
}

// ModelEvent is a streamed event whose payload is determined by its kind.
type ModelEvent struct {
	Kind     ModelEventKind
	Text     string
	ToolCall *ToolCall
}

// Validate returns an error when the payload does not match the event kind.
func (e ModelEvent) Validate() error {
	switch e.Kind {
	case EventTextDelta:
		if e.Text == "" {
			return fmt.Errorf("missing text for text delta event")
		}
		if e.ToolCall != nil {
			return fmt.Errorf("unexpected tool call for text delta event")
		}
	case EventToolCall:
		if e.Text != "" {
			return fmt.Errorf("unexpected text for tool call event")
		}
		if e.ToolCall == nil {
			return fmt.Errorf("missing tool call for tool call event")
		}
		if err := e.ToolCall.Validate(); err != nil {
			return fmt.Errorf("invalid tool call for tool call event: %w", err)
		}
	case EventCompleted:
		if e.Text != "" {
			return fmt.Errorf("unexpected text for completed event")
		}
		if e.ToolCall != nil {
			return fmt.Errorf("unexpected tool call for completed event")
		}
	default:
		return fmt.Errorf("unsupported event kind: %s", e.Kind)
	}

	return nil
}
