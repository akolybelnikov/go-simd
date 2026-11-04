package labworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/catalog"
	"github.com/akolybelnikov/go-simd/internal/evidence"
	"github.com/akolybelnikov/go-simd/internal/labapi"
	"github.com/akolybelnikov/go-simd/internal/labrun"
	"github.com/akolybelnikov/go-simd/internal/platform"
	"github.com/akolybelnikov/go-simd/internal/result"
)

func workerRequest(t *testing.T, jobID string) Request {
	t.Helper()
	plan, err := labapi.ResolveExperimentPlan(labapi.PlanRequest{
		Group: "addition", Variants: []string{catalog.Scalar}, Sizes: []int{8}, Count: 1, Benchtime: "1x",
	}, labrun.Runtime(platform.Current()), labapi.DefaultPlanPolicy())
	if err != nil {
		t.Fatal(err)
	}
	return Request{Schema: RequestSchema, JobID: jobID, Plan: plan}
}

func helperRunner(t *testing.T, mode string) *Runner {
	t.Helper()
	runner, err := NewRunnerWithCommand(func(ctx context.Context) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWorkerHelper$")
		cmd.Env = append(os.Environ(), "GO_SIMD_WORKER_HELPER="+mode)
		return cmd
	})
	if err != nil {
		t.Fatal(err)
	}
	return runner
}

func TestRunnerCompletesSelfWorker(t *testing.T) {
	got, err := helperRunner(t, "serve").Run(context.Background(), t.TempDir(), workerRequest(t, "complete-job"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Manifest.Status != "complete" || got.Capture == nil || len(got.Capture.Run.Observations) != 1 {
		t.Fatalf("result = %+v", got)
	}
	if got.Manifest.StartedAt.IsZero() || got.Manifest.FinishedAt.IsZero() {
		t.Fatalf("lifecycle timestamps missing: %+v", got.Manifest)
	}
}

func TestRunnerRefusesExistingJobDirectory(t *testing.T) {
	root := t.TempDir()
	request := workerRequest(t, "same-job")
	if _, err := helperRunner(t, "serve").Run(context.Background(), root, request); err != nil {
		t.Fatal(err)
	}
	if _, err := helperRunner(t, "serve").Run(context.Background(), root, request); err == nil || !strings.Contains(err.Error(), "artifact directory") {
		t.Fatalf("duplicate job error = %v", err)
	}
}

func TestRunnerPreservesCheckpointAfterWorkerCrash(t *testing.T) {
	got, err := helperRunner(t, "crash").Run(context.Background(), t.TempDir(), workerRequest(t, "crash-job"))
	if err == nil {
		t.Fatal("worker crash accepted")
	}
	if got.Manifest.Status != "failed" || got.Capture == nil || len(got.Capture.Run.Observations) != 1 {
		t.Fatalf("checkpoint not retained: %+v, err %v", got, err)
	}
	if _, statErr := os.Stat(filepath.Join(got.Directory, "measurements.json")); statErr != nil {
		t.Fatalf("checkpoint artifact missing: %v", statErr)
	}
}

func TestRunnerHardCancelsWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	got, err := helperRunner(t, "hang").Run(ctx, t.TempDir(), workerRequest(t, "cancel-job"))
	if !errors.Is(err, context.DeadlineExceeded) || got.Manifest.Status != "canceled" {
		t.Fatalf("status=%s err=%v", got.Manifest.Status, err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatal("worker was not killed promptly")
	}
}

func TestRecoverFinalizesInterruptedManifest(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "recover-job")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{Schema: manifestSchema, JobID: "recover-job", PlanID: "sha256:test", Status: "running", QueuedAt: time.Now().UTC()}
	if err := writeJSONAtomic(filepath.Join(directory, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	recovered, err := Recover(root)
	if err != nil || len(recovered) != 1 || recovered[0].Status != "failed" || recovered[0].FinishedAt.IsZero() {
		t.Fatalf("recovered=%+v err=%v", recovered, err)
	}
}

func TestPruneRemovesOnlyExpiredOrExcessTerminalJobs(t *testing.T) {
	root := t.TempDir()
	now := time.Now().UTC()
	writeManifest := func(id, status string, finished time.Time) {
		t.Helper()
		directory := filepath.Join(root, id)
		if err := os.Mkdir(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		manifest := Manifest{Schema: manifestSchema, JobID: id, Status: status, QueuedAt: finished, FinishedAt: finished}
		if err := writeJSONAtomic(filepath.Join(directory, "manifest.json"), manifest); err != nil {
			t.Fatal(err)
		}
	}
	writeManifest("newest", "complete", now.Add(-time.Minute))
	writeManifest("excess", "failed", now.Add(-2*time.Minute))
	writeManifest("expired", "canceled", now.Add(-48*time.Hour))
	writeManifest("active", "running", now.Add(-72*time.Hour))
	removed, err := Prune(root, RetentionPolicy{MaxJobs: 1, MaxAge: 24 * time.Hour}, now)
	if err != nil || !slices.Equal(removed, []string{"excess", "expired"}) {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
	for _, id := range []string{"newest", "active"} {
		if _, err := os.Stat(filepath.Join(root, id)); err != nil {
			t.Fatalf("preserved job %s missing: %v", id, err)
		}
	}
}

func TestCaptureArtifactSizeIsBounded(t *testing.T) {
	capture := evidence.Capture{
		Schema: evidence.Schema, Status: "complete", Group: "addition",
		Run: result.Run{ID: "large", Metadata: strings.Repeat("x", MaxArtifactBytes)},
	}
	if err := writeCaptureAtomic(filepath.Join(t.TempDir(), "measurements.json"), capture); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized capture error = %v", err)
	}
}

// TestWorkerHelper is invoked as a child test binary by the parent-side tests.
func TestWorkerHelper(t *testing.T) {
	mode := os.Getenv("GO_SIMD_WORKER_HELPER")
	if mode == "" {
		return
	}
	var request Request
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	encoder := json.NewEncoder(os.Stdout)
	switch mode {
	case "serve":
		data, _ := json.Marshal(request)
		if err := Serve(context.Background(), bytes.NewReader(data), os.Stdout, labapi.DefaultPlanPolicy()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	case "crash":
		_ = encoder.Encode(Event{Schema: EventSchema, Type: "started", PlanID: request.Plan.ID})
		capture := evidence.Capture{
			Schema: evidence.Schema, Status: "failed", Error: "worker crashed", Group: request.Plan.Group,
			Run: result.Run{ID: request.JobID, Observations: []result.Observation{{Repetition: 1}}},
		}
		_ = encoder.Encode(Event{Schema: EventSchema, Type: "checkpoint", PlanID: request.Plan.ID, Capture: &capture})
		os.Exit(7)
	case "hang":
		_ = encoder.Encode(Event{Schema: EventSchema, Type: "started", PlanID: request.Plan.ID})
		time.Sleep(30 * time.Second)
		os.Exit(0)
	default:
		os.Exit(3)
	}
}
