// Package agent owns the provider-neutral conversation loop.
package agent

import (
	"context"
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

// Validate returns an error when the message is structurally invalid.
func (m Message) Validate() error {
	switch m.Role {
	case RoleUser:
		if m.Content == "" {
			return fmt.Errorf("missing content for user message")
		}
		if len(m.ToolCalls) > 0 {
			return fmt.Errorf("unexpected tool calls for user message")
		}
		if m.ToolCallID != "" {
			return fmt.Errorf("unexpected tool call ID for user message")
		}
	case RoleAssistant:
		if m.Content == "" && len(m.ToolCalls) == 0 {
			return fmt.Errorf("missing content or tool calls for assistant message")
		}
		if m.ToolCallID != "" {
			return fmt.Errorf("unexpected tool call ID for assistant message")
		}
		for _, call := range m.ToolCalls {
			if err := call.Validate(); err != nil {
				return fmt.Errorf("invalid tool call for assistant message: %w", err)
			}
		}
	case RoleTool:
		if m.Content == "" {
			return fmt.Errorf("missing content for tool message")
		}
		if len(m.ToolCalls) > 0 {
			return fmt.Errorf("unexpected tool calls for tool message")
		}
		if m.ToolCallID == "" {
			return fmt.Errorf("missing tool call ID for tool message")
		}
	default:
		return fmt.Errorf("unsupported message role: %s", m.Role)
	}

	return nil
}

// ToolDefinition is a description of a tool for an LLM call.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// Validate returns an error when the tool definition is structurally invalid.
func (t ToolDefinition) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("missing tool name")
	}
	if t.Description == "" {
		return fmt.Errorf("missing tool description")
	}

	var inputSchema map[string]json.RawMessage
	if err := json.Unmarshal(t.InputSchema, &inputSchema); err != nil {
		return fmt.Errorf("invalid input schema: %w", err)
	}
	if inputSchema == nil {
		return fmt.Errorf("missing input schema")
	}
	val, ok := inputSchema["type"]
	if !ok {
		return fmt.Errorf("missing 'type' in input schema")
	}

	var objType string
	if err := json.Unmarshal(val, &objType); err != nil {
		return fmt.Errorf("invalid 'type' in input schema: %w", err)
	}
	if objType != "object" {
		return fmt.Errorf("input schema 'type' must be 'object', got: %s", objType)
	}

	return nil
}

// validateToolDefinitions returns an error if any tool definitions are invalid or if there are duplicate tool names.
func validateToolDefinitions(definitions []ToolDefinition) error {
	d := make(map[string]int, len(definitions))
	for i, def := range definitions {
		if idx, ok := d[def.Name]; ok {
			return fmt.Errorf("found a duplicate tool name at the index %d (previously defined at the index %d): %s", i, idx, def.Name)
		}
		if err := def.Validate(); err != nil {
			return fmt.Errorf("found an invalid definition at the index %d: %w", i, err)
		}
		d[def.Name] = i
	}

	return nil
}

// ModelRequest is a request to the model with trusted instructions and the ordered conversation history.
type ModelRequest struct {
	Instructions string
	Messages     []Message
	Tools        []ToolDefinition
}

// Validate returns an error when the request is structurally invalid.
func (r ModelRequest) Validate() error {
	if r.Instructions == "" {
		return fmt.Errorf("missing instructions")
	}

	if len(r.Messages) == 0 {
		return fmt.Errorf("missing messages")
	}
	for i, msg := range r.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("the request contains an invalid message at the index %d: %w", i, err)
		}
	}

	if err := validateToolDefinitions(r.Tools); err != nil {
		return fmt.Errorf("the request contains invalid tool definitions: %w", err)
	}

	return nil
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

// ModelStream provides events for one model interaction.
type ModelStream interface {
	// Next returns the next event and returns io.EOF when the stream ends normally.
	Next() (ModelEvent, error)
	// Close releases stream resources and may be called before the stream is drained.
	Close() error
}

// ModelClient is an interface for a client that can stream model events.
type ModelClient interface {
	// Stream opens one model stream.
	Stream(ctx context.Context, req ModelRequest) (ModelStream, error)
}
