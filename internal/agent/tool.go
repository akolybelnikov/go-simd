package agent

import "context"

// ToolDispatcher dispatches one validated, registered tool call and returns its tool-result message.
type ToolDispatcher interface {
	Dispatch(ctx context.Context, toolCall ToolCall) (Message, error)
}
