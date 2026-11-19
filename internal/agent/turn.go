package agent

import (
	"context"
	"fmt"
)

const (
	maxTurnMessages = 64
	maxTurnBytes    = 512 << 10
)

// TurnLimits bounds model requests and tool dispatches within a single turn.
type TurnLimits struct {
	MaxModelSteps int
	MaxToolCalls  int
}

// RunTurn advances the conversation until the model answers without tool calls.
// The caller supplies a context with a deadline and a dispatcher for the tools
// advertised by this runtime. Errors terminate the turn.
func (r *Runtime) RunTurn(ctx context.Context, messages []Message, dispatcher ToolDispatcher, limits TurnLimits) (Message, error) {
	if ctx == nil {
		return Message{}, fmt.Errorf("missing turn context")
	}
	if _, ok := ctx.Deadline(); !ok {
		return Message{}, fmt.Errorf("turn context requires a deadline")
	}
	if limits.MaxModelSteps < 1 || limits.MaxToolCalls < 0 {
		return Message{}, fmt.Errorf("invalid turn limits")
	}
	if err := validateTurnHistory(messages, false); err != nil {
		return Message{}, err
	}
	history := cloneMessages(messages)
	remainingCalls := limits.MaxToolCalls
	for step := 1; step <= limits.MaxModelSteps; step++ {
		if err := ctx.Err(); err != nil {
			return Message{}, err
		}
		reply, err := r.RunModelStep(ctx, history)
		if err != nil {
			return Message{}, fmt.Errorf("model step %d: %w", step, err)
		}
		if err := ctx.Err(); err != nil {
			return Message{}, err
		}
		if len(reply.ToolCalls) == 0 {
			return reply, nil
		}
		if step == limits.MaxModelSteps {
			return Message{}, fmt.Errorf("model step limit reached before final answer")
		}
		if len(reply.ToolCalls) > remainingCalls {
			return Message{}, fmt.Errorf("tool call limit exceeded")
		}
		if dispatcher == nil {
			return Message{}, fmt.Errorf("missing tool dispatcher")
		}
		history = append(history, reply)
		if err := validateTurnHistory(history, true); err != nil {
			return Message{}, err
		}
		if len(history)+len(reply.ToolCalls) > maxTurnMessages {
			return Message{}, fmt.Errorf("turn history message limit exceeded")
		}
		remainingCalls -= len(reply.ToolCalls)
		for _, call := range reply.ToolCalls {
			if err := ctx.Err(); err != nil {
				return Message{}, err
			}
			call.Arguments = call.Arguments.Clone()
			result, err := dispatcher.Dispatch(ctx, call)
			if err != nil {
				return Message{}, fmt.Errorf("dispatch %q: %w", call.Name, err)
			}
			if err := result.Validate(); err != nil {
				return Message{}, fmt.Errorf("invalid tool result: %w", err)
			}
			if result.Role != RoleTool || result.ToolCallID != call.ID {
				return Message{}, fmt.Errorf("tool result does not match call %q", call.ID)
			}
			history = append(history, result)
			if err := validateTurnHistory(history, true); err != nil {
				return Message{}, err
			}
		}
	}
	return Message{}, fmt.Errorf("model step limit reached")
}

// validateTurnHistory can permit a pending batch at the end. A later model
// request can reuse IDs after the results for the preceding batch are complete.
func validateTurnHistory(messages []Message, allowPending bool) error {
	if len(messages) > maxTurnMessages {
		return fmt.Errorf("turn history message limit exceeded")
	}
	remaining := maxTurnBytes
	pending := make(map[string]bool)
	for _, msg := range messages {
		remaining -= len(msg.Content) + len(msg.ToolCallID)
		if remaining < 0 {
			return fmt.Errorf("turn history byte limit exceeded")
		}
		if msg.Role == RoleTool {
			if !pending[msg.ToolCallID] {
				return fmt.Errorf("unexpected tool result for %q", msg.ToolCallID)
			}
			delete(pending, msg.ToolCallID)
			continue
		}
		if len(pending) != 0 {
			return fmt.Errorf("tool results are missing before the next message")
		}
		for _, call := range msg.ToolCalls {
			remaining -= len(call.ID) + len(call.Name) + len(call.Arguments)
			if remaining < 0 {
				return fmt.Errorf("turn history byte limit exceeded")
			}
			if pending[call.ID] {
				return fmt.Errorf("duplicate tool call ID %q", call.ID)
			}
			pending[call.ID] = true
		}
	}
	if !allowPending && len(pending) != 0 {
		return fmt.Errorf("tool results are missing at the end of history")
	}
	return nil
}
