package ollama

import (
	"encoding/json"
)

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Think    bool          `json:"think"`
	Tools    []chatTool    `json:"tools,omitempty"`
}

type chatMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Thinking  string         `json:"thinking,omitempty"`
	ToolCalls []chatToolCall `json:"tool_calls,omitempty"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
}

type chatTool struct {
	Type     string           `json:"type"`
	Function chatToolFunction `json:"function"`
}

type chatToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type chatToolCall struct {
	Type     string               `json:"type"`
	Function chatToolCallFunction `json:"function"`
}

type chatToolCallFunction struct {
	Index     int             `json:"index"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
