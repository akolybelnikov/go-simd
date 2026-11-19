package labdispatch

import (
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/agent"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labtools"
)

const planToolName = "plan_experiment"

// ExperimentPlanTool describes the proposal capability used by the CLI.
func ExperimentPlanTool() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        planToolName,
		Description: "Propose one SIMD benchmark with an exact group, strategy IDs, element sizes, repetition count, and Go benchtime. Uses the lab's default generated fixture. The user reviews and confirms the resolved plan separately.",
		InputSchema: []byte(`{"type":"object","additionalProperties":false,"required":["group","variants","sizes","count","benchtime"],"properties":{"group":{"type":"string"},"variants":{"type":"array","items":{"type":"string"}},"sizes":{"type":"array","items":{"type":"integer"}},"count":{"type":"integer"},"benchtime":{"type":"string"}}}`),
	}
}

// ProposeExperiment validates a model proposal and returns the canonical plan
// for review by the application.
func ProposeExperiment(authority labtools.ModelAuthority, call agent.ToolCall) (labapi.ExperimentPlan, error) {
	if call.Name != planToolName {
		return labapi.ExperimentPlan{}, fmt.Errorf("unexpected experiment tool %q", call.Name)
	}
	if err := call.Validate(); err != nil {
		return labapi.ExperimentPlan{}, err
	}
	var args struct {
		Group     string   `json:"group"`
		Variants  []string `json:"variants"`
		Sizes     []int    `json:"sizes"`
		Count     int      `json:"count"`
		Benchtime string   `json:"benchtime"`
	}
	if err := decodeArguments(call.Arguments, &args); err != nil {
		return labapi.ExperimentPlan{}, fmt.Errorf("decode experiment proposal: %w", err)
	}
	return authority.ProposeExperiment(labapi.PlanRequest{Group: args.Group, Variants: args.Variants, Sizes: args.Sizes, Count: args.Count, Benchtime: args.Benchtime})
}
