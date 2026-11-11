package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Runtime coordinates model interactions using fixed instructions and allowed tools.
type Runtime struct {
	client       ModelClient
	instructions string
	tools        []ToolDefinition
}

// modelRequest builds and validates a model request from the given runtime.
func (r *Runtime) modelRequest(msgs []Message) (ModelRequest, error) {
	if r == nil {
		return ModelRequest{}, fmt.Errorf("nil runtime receiver")
	}

	req := ModelRequest{
		Instructions: r.instructions,
		Messages:     cloneMessages(msgs),
		Tools:        cloneToolDefinitions(r.tools),
	}

	if err := req.Validate(); err != nil {
		return ModelRequest{}, fmt.Errorf("invalid model request in runtime: %w", err)
	}

	return req, nil
}

func (r *Runtime) isToolRegistered(name string) bool {
	for _, tool := range r.tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

// RunModelStep consumes one stream interaction with an LLM.
func (r *Runtime) RunModelStep(ctx context.Context, msgs []Message) (Message, error) {
	mRequest, err := r.modelRequest(msgs)
	if err != nil {
		return Message{}, err
	}

	stream, err := r.client.Stream(ctx, mRequest)
	if err != nil {
		return Message{}, fmt.Errorf("stream error: %w", err)
	}
	defer stream.Close()

	bd := strings.Builder{}
	completed := false
	var toolCalls []ToolCall

	for {
		event, err := stream.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !completed {
					return Message{}, fmt.Errorf("model stream ended without completion")
				}
				break
			}
			return Message{}, fmt.Errorf("unexpected stream error: %w", err)
		}

		if err := event.Validate(); err != nil {
			return Message{}, fmt.Errorf("unexpected event error: %w", err)
		}

		if completed {
			return Message{}, fmt.Errorf("received an event after completion")
		}

		switch event.Kind {
		case EventCompleted:
			completed = true
		case EventTextDelta:
			_, _ = bd.WriteString(event.Text)
		case EventToolCall:
			tc := (*event.ToolCall)
			if !r.isToolRegistered(tc.Name) {
				return Message{}, fmt.Errorf("received an unregistered tool call: %q", tc.Name)
			}
			tc.Arguments = bytes.Clone(event.ToolCall.Arguments)
			toolCalls = append(toolCalls, tc)
		default:
		}

	}

	msg := Message{Role: RoleAssistant, Content: bd.String(), ToolCalls: toolCalls}
	if err := msg.Validate(); err != nil {
		return Message{}, fmt.Errorf("invalid model response: %w", err)
	}

	return msg, nil
}

// NewRuntime creates a new runtime with the given arguments.
func NewRuntime(client ModelClient, instructions string, tools []ToolDefinition) (*Runtime, error) {
	if client == nil {
		return nil, fmt.Errorf("missing model client")
	}
	if instructions == "" {
		return nil, fmt.Errorf("missing instructions")
	}

	if err := validateToolDefinitions(tools); err != nil {
		return nil, fmt.Errorf("invalid tool definitions in runtime: %w", err)
	}

	return &Runtime{
		client:       client,
		instructions: instructions,
		tools:        cloneToolDefinitions(tools),
	}, nil
}

// cloneToolDefinitions creates a deep copy of the given tool definitions to avoid mutating the input schema.
func cloneToolDefinitions(defs []ToolDefinition) []ToolDefinition {
	cp := slices.Clone(defs)
	for i, def := range cp {
		cp[i].InputSchema = bytes.Clone(def.InputSchema)
	}

	return cp
}

// cloneMessages creates a deep copy of the given messages to avoid mutating tool calls arguments.
func cloneMessages(msgs []Message) []Message {
	cp := slices.Clone(msgs)
	for i := range cp {
		cp[i].ToolCalls = slices.Clone(msgs[i].ToolCalls)
		for j := range cp[i].ToolCalls {
			cp[i].ToolCalls[j].Arguments = bytes.Clone(msgs[i].ToolCalls[j].Arguments)
		}
	}

	return cp
}
