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

	for _, msg := range req.Messages {
		switch msg.Role {
		case agent.RoleUser:
			msgs = append(msgs, chatMessage{
				Role:    userRole,
				Content: msg.Content,
			})
		case agent.RoleAssistant:
			if len(msg.ToolCalls) > 0 {
				return nil, fmt.Errorf("tools are not supported yet")
			}
			msgs = append(msgs, chatMessage{
				Role:    assistantRole,
				Content: msg.Content,
			})
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
