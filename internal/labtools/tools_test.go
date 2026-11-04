package labtools

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labworker"
)

type recordingExecutor struct {
	mu       sync.Mutex
	requests []labworker.Request
}

func (e *recordingExecutor) Execute(_ context.Context, request labworker.Request) (labworker.RunResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requests = append(e.requests, request)
	return labworker.RunResult{Manifest: labworker.Manifest{JobID: request.JobID, PlanID: request.Plan.ID, Status: "complete"}}, nil
}

func toolRuntime() labapi.RuntimeContext {
	return labapi.RuntimeContext{GoVersion: "go-test", OS: "test", Arch: "test", VectorBits: 256, Float32Lanes: 8, GOMAXPROCS: 4}
}

func proposal() labapi.PlanRequest {
	return labapi.PlanRequest{Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 3, Benchtime: "1x"}
}

func TestOnlyConfirmedStoredPlanCanExecute(t *testing.T) {
	executor := &recordingExecutor{}
	tools, err := newWithRandom(toolRuntime(), labapi.DefaultPlanPolicy(), executor, bytes.NewReader(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := tools.ProposeExperiment(proposal())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tools.ExecuteConfirmed(context.Background(), plan.ID); err == nil || !strings.Contains(err.Error(), "not confirmed") {
		t.Fatalf("unconfirmed execution error = %v", err)
	}
	if _, err := tools.ExecuteConfirmed(context.Background(), "sha256:unknown"); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown execution error = %v", err)
	}
	if err := tools.ConfirmExperiment(plan.ID); err != nil {
		t.Fatal(err)
	}
	result, err := tools.ExecuteConfirmed(context.Background(), plan.ID)
	if err != nil || result.Manifest.Status != "complete" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := tools.ExecuteConfirmed(context.Background(), plan.ID); err == nil || !strings.Contains(err.Error(), "finished") {
		t.Fatalf("replay error = %v", err)
	}
	if len(executor.requests) != 1 || executor.requests[0].Plan.ID != plan.ID || executor.requests[0].JobID != "job-00000000000000000000000000000000" {
		t.Fatalf("requests = %+v", executor.requests)
	}
}

func TestReturnedPlanMutationCannotBroadenStoredAuthority(t *testing.T) {
	executor := &recordingExecutor{}
	tools, _ := newWithRandom(toolRuntime(), labapi.DefaultPlanPolicy(), executor, bytes.NewReader(make([]byte, 16)))
	plan, err := tools.ProposeExperiment(proposal())
	if err != nil {
		t.Fatal(err)
	}
	plan.Cases[0].Elements = 1 << 30
	plan.Cases[0].Name = "attacker-controlled"
	if err := tools.ConfirmExperiment(plan.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := tools.ExecuteConfirmed(context.Background(), plan.ID); err != nil {
		t.Fatal(err)
	}
	got := executor.requests[0].Plan.Cases[0]
	if got.Elements != 8 || got.Name == "attacker-controlled" {
		t.Fatalf("stored plan was mutated: %+v", got)
	}
}

func TestProposalLimitsAndSessionBound(t *testing.T) {
	tools, _ := newWithRandom(toolRuntime(), labapi.DefaultPlanPolicy(), &recordingExecutor{}, bytes.NewReader(make([]byte, 16)))
	request := proposal()
	request.Count = labapi.DefaultPlanPolicy().MaxCount + 1
	if _, err := tools.ProposeExperiment(request); err == nil {
		t.Fatal("oversized proposal accepted")
	}
}
