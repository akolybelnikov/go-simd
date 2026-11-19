package agent_test

import (
	"context"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/stretchr/testify/require"
)

type modelFunc func(context.Context, agent.ModelRequest) (agent.ModelStream, error)

func (f modelFunc) Stream(ctx context.Context, req agent.ModelRequest) (agent.ModelStream, error) {
	return f(ctx, req)
}

type dispatchFunc func(context.Context, agent.ToolCall) (agent.Message, error)

func (f dispatchFunc) Dispatch(ctx context.Context, call agent.ToolCall) (agent.Message, error) {
	return f(ctx, call)
}

func TestRunTurn(t *testing.T) {
	for _, tt := range []struct {
		name         string
		limits       agent.TurnLimits
		cancel       bool
		wrongID      bool
		duplicate    bool
		wantError    string
		wantSteps    int
		wantDispatch int
	}{
		{name: "answer uses tool result", limits: agent.TurnLimits{MaxModelSteps: 2, MaxToolCalls: 1}, wantSteps: 2, wantDispatch: 1},
		{name: "no step left to read result", limits: agent.TurnLimits{MaxModelSteps: 1, MaxToolCalls: 1}, wantError: "model step limit", wantSteps: 1},
		{name: "tool budget exhausted", limits: agent.TurnLimits{MaxModelSteps: 2}, wantError: "tool call limit", wantSteps: 1},
		{name: "mismatched result", limits: agent.TurnLimits{MaxModelSteps: 2, MaxToolCalls: 1}, wrongID: true, wantError: "does not match", wantSteps: 1, wantDispatch: 1},
		{name: "duplicate calls", limits: agent.TurnLimits{MaxModelSteps: 2, MaxToolCalls: 2}, duplicate: true, wantError: "duplicate tool call ID", wantSteps: 1},
		{name: "cancelled", limits: agent.TurnLimits{MaxModelSteps: 2, MaxToolCalls: 1}, cancel: true, wantError: "context canceled"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			if tt.cancel {
				cancel()
			}
			user := agent.Message{Role: agent.RoleUser, Content: "Which addition strategies are available?"}
			call := agent.ToolCall{ID: "call-1", Name: "explore_catalog", Arguments: []byte(`{"group":"addition"}`)}
			result := agent.Message{Role: agent.RoleTool, ToolCallID: call.ID, Content: `{"strategies":["scalar"]}`}
			steps, dispatches := 0, 0
			model := modelFunc(func(_ context.Context, req agent.ModelRequest) (agent.ModelStream, error) {
				steps++
				event := agent.ModelEvent{Kind: agent.EventToolCall, ToolCall: &call}
				if steps == 1 {
					require.Equal(t, []agent.Message{user}, req.Messages)
				} else {
					require.Equal(t, []agent.Message{user, {Role: agent.RoleAssistant, ToolCalls: []agent.ToolCall{call}}, result}, req.Messages)
					event = agent.ModelEvent{Kind: agent.EventTextDelta, Text: "Scalar is available."}
				}
				events := []agent.ModelEvent{event}
				if tt.duplicate {
					events = append(events, event)
				}
				return &fakeModelStream{events: append(events, agent.ModelEvent{Kind: agent.EventCompleted})}, nil
			})
			dispatch := dispatchFunc(func(_ context.Context, requested agent.ToolCall) (agent.Message, error) {
				dispatches++
				require.Equal(t, call, requested)
				if tt.wrongID {
					return agent.Message{Role: agent.RoleTool, ToolCallID: "other", Content: result.Content}, nil
				}
				return result, nil
			})
			runtime, err := agent.NewRuntime(model, "Use the catalog.", []agent.ToolDefinition{{Name: call.Name, Description: "Read the catalog", InputSchema: []byte(`{"type":"object"}`)}})
			require.NoError(t, err)
			reply, err := runtime.RunTurn(ctx, []agent.Message{user}, dispatch, tt.limits)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				require.Zero(t, reply)
			} else {
				require.NoError(t, err)
				require.Equal(t, "Scalar is available.", reply.Content)
			}
			require.Equal(t, tt.wantSteps, steps)
			require.Equal(t, tt.wantDispatch, dispatches)
		})
	}
}
