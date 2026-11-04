// Package labworker defines the private self-worker protocol and parent-side
// artifact lifecycle for isolated benchmark execution.
package labworker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
)

const (
	RequestSchema = "go-simd.worker-request/v1"
	EventSchema   = "go-simd.worker-event/v1"
	MaxRequest    = 1 << 20
	MaxEvent      = 4 << 20
)

type Request struct {
	Schema string                `json:"schema"`
	JobID  string                `json:"job_id"`
	Plan   labapi.ExperimentPlan `json:"plan"`
}

type Event struct {
	Schema  string            `json:"schema"`
	Type    string            `json:"type"`
	PlanID  string            `json:"plan_id"`
	Capture *evidence.Capture `json:"capture,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// Serve handles exactly one request. Events are newline-delimited so the
// parent can persist checkpoints before the process reaches final status.
func Serve(ctx context.Context, in io.Reader, out io.Writer, policy labapi.PlanPolicy) error {
	data, err := io.ReadAll(io.LimitReader(in, MaxRequest+1))
	if err != nil {
		return fmt.Errorf("read worker request: %w", err)
	}
	if len(data) > MaxRequest {
		return fmt.Errorf("worker request exceeds %d bytes", MaxRequest)
	}
	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		return fmt.Errorf("decode worker request: %w", err)
	}
	if request.Schema != RequestSchema || request.JobID == "" {
		return fmt.Errorf("invalid worker request header")
	}
	encoder := json.NewEncoder(out)
	emit := func(event Event) error {
		event.Schema = EventSchema
		event.PlanID = request.Plan.ID
		return encoder.Encode(event)
	}
	if err := emit(Event{Type: "started"}); err != nil {
		return fmt.Errorf("write worker start: %w", err)
	}
	capture, runErr := labrun.Execute(ctx, request.JobID, request.Plan, policy, func(c evidence.Capture) error {
		return emit(Event{Type: "checkpoint", Capture: &c})
	})
	if capture.Schema == "" && runErr != nil {
		capture = evidence.Capture{
			Schema: evidence.Schema, Status: "failed", Error: runErr.Error(),
			Group: request.Plan.Group,
		}
	}
	final := Event{Type: "final", Capture: &capture}
	if runErr != nil {
		final.Error = runErr.Error()
	}
	if err := emit(final); err != nil {
		return fmt.Errorf("write worker final: %w", err)
	}
	return runErr
}
