package ollama

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/agent"
)

const (
	systemRole    = "system"
	userRole      = "user"
	assistantRole = "assistant"
)

func (c *Client) buildChatRequest(req agent.ModelRequest) (*chatRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid model request: %w", err)
	}
	msgs := make([]chatMessage, 0, len(req.Messages)+1)
	msgs = append(msgs, chatMessage{
		Role:    systemRole,
		Content: req.Instructions,
	})

	pendingToolCalls := make(map[string]string)

	for _, msg := range req.Messages {
		switch msg.Role {
		case agent.RoleUser:
			msgs = append(msgs, chatMessage{
				Role:    userRole,
				Content: msg.Content,
			})
		case agent.RoleAssistant:
			chatMsg := chatMessage{
				Role:    assistantRole,
				Content: msg.Content,
			}
			for idx, tc := range msg.ToolCalls {
				chatMsg.ToolCalls = append(chatMsg.ToolCalls, chatToolCall{
					Type: "function",
					Function: chatToolCallFunction{
						Index:     idx,
						Name:      tc.Name,
						Arguments: tc.Arguments.Clone(),
					},
				})
				pendingToolCalls[tc.ID] = tc.Name
			}
			msgs = append(msgs, chatMsg)
		case agent.RoleTool:
			toolName, ok := pendingToolCalls[msg.ToolCallID]
			if !ok {
				return nil, fmt.Errorf("unmatched tool call ID %q", msg.ToolCallID)
			}
			msgs = append(msgs, chatMessage{
				Role:     "tool",
				ToolName: toolName,
				Content:  msg.Content,
			})
			delete(pendingToolCalls, msg.ToolCallID)
		default:
			return nil, fmt.Errorf("unsupported message role: %s", msg.Role)
		}
	}

	tools := make([]chatTool, 0, len(req.Tools))
	for _, tool := range req.Tools {
		tools = append(tools, chatTool{
			Type: "function",
			Function: chatToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	return &chatRequest{
		Model:    c.modelName,
		Messages: msgs,
		Stream:   true,
		Think:    true,
		Tools:    tools,
	}, nil
}
