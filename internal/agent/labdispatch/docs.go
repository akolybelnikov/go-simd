package labdispatch

import (
	"encoding/json"
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/labdocs"
)

const docsToolName = "search_lab_docs"

// DocsTool returns the definition for the curated documentation lookup.
func DocsTool() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        docsToolName,
		Description: "Find curated go-simd documentation notes and their source sections by keywords.",
		InputSchema: []byte(`{"type":"object","properties":{"query":{"type":"string","minLength":1,"maxLength":256}},"required":["query"],"additionalProperties":false}`),
	}
}

func searchDocs(call agent.ToolCall) (agent.Message, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := decodeArguments(call.Arguments, &args); err != nil {
		return agent.Message{}, fmt.Errorf("invalid documentation arguments: %w", err)
	}
	results, err := labdocs.Search(args.Query)
	if err != nil {
		return agent.Message{}, err
	}
	content, err := json.Marshal(results)
	if err != nil {
		return agent.Message{}, fmt.Errorf("encode documentation results: %w", err)
	}
	return agent.Message{Role: agent.RoleTool, ToolCallID: call.ID, Content: string(content)}, nil
}
