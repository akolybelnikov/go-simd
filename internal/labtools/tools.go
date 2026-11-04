// Package labtools exposes the narrow authority intended for an agent model.
// It deliberately has no shell, path, URL, or raw worker-argument capability.
package labtools

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labworker"
)

const maxSessionPlans = 16

type Executor interface {
	Execute(context.Context, labworker.Request) (labworker.RunResult, error)
}

type planState string

const (
	proposed  planState = "proposed"
	confirmed planState = "confirmed"
	running   planState = "running"
	finished  planState = "finished"
)

type storedPlan struct {
	plan  labapi.ExperimentPlan
	state planState
}

type Toolset struct {
	mu       sync.Mutex
	runtime  labapi.RuntimeContext
	policy   labapi.PlanPolicy
	executor Executor
	random   io.Reader
	plans    map[string]storedPlan
}

// ModelAuthority is the complete model-facing surface. Confirmation is
// intentionally absent and remains an application/user-control action.
type ModelAuthority struct{ tools *Toolset }

func (t *Toolset) ModelAuthority() ModelAuthority { return ModelAuthority{tools: t} }

func (a ModelAuthority) ProposeExperiment(request labapi.PlanRequest) (labapi.ExperimentPlan, error) {
	return a.tools.ProposeExperiment(request)
}

func (a ModelAuthority) ExecuteConfirmed(ctx context.Context, planID string) (labworker.RunResult, error) {
	return a.tools.ExecuteConfirmed(ctx, planID)
}

func New(runtime labapi.RuntimeContext, policy labapi.PlanPolicy, executor Executor) (*Toolset, error) {
	return newWithRandom(runtime, policy, executor, rand.Reader)
}

func newWithRandom(runtime labapi.RuntimeContext, policy labapi.PlanPolicy, executor Executor, random io.Reader) (*Toolset, error) {
	if executor == nil {
		return nil, fmt.Errorf("lab executor is required")
	}
	if random == nil {
		return nil, fmt.Errorf("job ID source is required")
	}
	return &Toolset{runtime: runtime, policy: policy, executor: executor, random: random, plans: make(map[string]storedPlan)}, nil
}

// ProposeExperiment is safe for model invocation: it only accepts the bounded,
// explicit PlanRequest contract and stores an immutable canonical copy.
func (t *Toolset) ProposeExperiment(request labapi.PlanRequest) (labapi.ExperimentPlan, error) {
	plan, err := labapi.ResolveExperimentPlan(request, t.runtime, t.policy)
	if err != nil {
		return labapi.ExperimentPlan{}, err
	}
	stored, err := clonePlan(plan)
	if err != nil {
		return labapi.ExperimentPlan{}, err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, exists := t.plans[plan.ID]; !exists && len(t.plans) >= maxSessionPlans {
		return labapi.ExperimentPlan{}, fmt.Errorf("session plan limit reached")
	}
	t.plans[plan.ID] = storedPlan{plan: stored, state: proposed}
	return clonePlan(plan)
}

// ConfirmExperiment is an application/user-control action, not an agent tool.
// A frontend calls it only after presenting the exact immutable plan to a user.
func (t *Toolset) ConfirmExperiment(planID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	stored, ok := t.plans[planID]
	if !ok {
		return fmt.Errorf("unknown experiment plan")
	}
	if stored.state != proposed {
		return fmt.Errorf("experiment plan is %s, not proposed", stored.state)
	}
	if err := labapi.RevalidateExperimentPlan(stored.plan, t.runtime, t.policy); err != nil {
		return err
	}
	stored.state = confirmed
	t.plans[planID] = stored
	return nil
}

// ExecuteConfirmed accepts only a stored plan ID. It creates the opaque job ID
// internally and permits one execution, so model text cannot broaden cases or
// select an artifact path between confirmation and execution.
func (t *Toolset) ExecuteConfirmed(ctx context.Context, planID string) (labworker.RunResult, error) {
	if ctx == nil {
		return labworker.RunResult{}, fmt.Errorf("nil execution context")
	}
	t.mu.Lock()
	stored, ok := t.plans[planID]
	if !ok {
		t.mu.Unlock()
		return labworker.RunResult{}, fmt.Errorf("unknown experiment plan")
	}
	if stored.state != confirmed {
		t.mu.Unlock()
		return labworker.RunResult{}, fmt.Errorf("experiment plan is %s, not confirmed", stored.state)
	}
	stored.state = running
	t.plans[planID] = stored
	t.mu.Unlock()

	jobID, err := t.newJobID()
	if err != nil {
		t.finish(planID)
		return labworker.RunResult{}, err
	}
	output, runErr := t.executor.Execute(ctx, labworker.Request{Schema: labworker.RequestSchema, JobID: jobID, Plan: stored.plan})
	t.finish(planID)
	return output, runErr
}

func (t *Toolset) finish(planID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	stored := t.plans[planID]
	stored.state = finished
	t.plans[planID] = stored
}

func (t *Toolset) newJobID() (string, error) {
	data := make([]byte, 16)
	if _, err := io.ReadFull(t.random, data); err != nil {
		return "", fmt.Errorf("create job ID: %w", err)
	}
	return "job-" + hex.EncodeToString(data), nil
}

func clonePlan(plan labapi.ExperimentPlan) (labapi.ExperimentPlan, error) {
	data, err := json.Marshal(plan)
	if err != nil {
		return labapi.ExperimentPlan{}, err
	}
	var clone labapi.ExperimentPlan
	if err := json.Unmarshal(data, &clone); err != nil {
		return labapi.ExperimentPlan{}, err
	}
	return clone, nil
}
