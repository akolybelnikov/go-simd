package labworker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/akolybelnikov/go-simd/internal/evidence"
)

const manifestSchema = "go-simd.worker-manifest/v1"
const MaxArtifactBytes = 4 << 20

var jobIDPattern = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$")

type Manifest struct {
	Schema       string    `json:"schema"`
	JobID        string    `json:"job_id"`
	PlanID       string    `json:"plan_id"`
	Status       string    `json:"status"`
	QueuedAt     time.Time `json:"queued_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	Observations int       `json:"observations"`
	Error        string    `json:"error,omitempty"`
}

type RunResult struct {
	Directory string
	Manifest  Manifest
	Capture   *evidence.Capture
}

type CommandFactory func(context.Context) *exec.Cmd

type Runner struct {
	command CommandFactory
}

func NewRunner(executable string, args ...string) (*Runner, error) {
	if executable == "" {
		return nil, fmt.Errorf("worker executable is required")
	}
	return NewRunnerWithCommand(func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, executable, args...)
	})
}

func NewRunnerWithCommand(command CommandFactory) (*Runner, error) {
	if command == nil {
		return nil, fmt.Errorf("worker command is required")
	}
	return &Runner{command: command}, nil
}

// Run starts one child process, persists every checkpoint, and always attempts
// to replace a running manifest with a terminal state before returning.
func (r *Runner) Run(ctx context.Context, artifactRoot string, request Request) (RunResult, error) {
	now := time.Now().UTC()
	manifest := Manifest{Schema: manifestSchema, JobID: request.JobID, PlanID: request.Plan.ID, Status: "queued", QueuedAt: now}
	result := RunResult{Manifest: manifest}
	if ctx == nil {
		return result, fmt.Errorf("nil worker context")
	}
	if request.Schema != RequestSchema || !jobIDPattern.MatchString(request.JobID) {
		return result, fmt.Errorf("invalid worker request or job ID")
	}
	if artifactRoot == "" {
		return result, fmt.Errorf("artifact root is required")
	}
	directory := filepath.Join(artifactRoot, request.JobID)
	result.Directory = directory
	if err := os.MkdirAll(artifactRoot, 0o700); err != nil {
		return result, fmt.Errorf("create artifact root: %w", err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		return result, fmt.Errorf("create artifact directory: %w", err)
	}
	manifestPath := filepath.Join(directory, "manifest.json")
	if err := writeJSONAtomic(manifestPath, manifest); err != nil {
		return result, err
	}
	requestData, err := json.Marshal(request)
	if err != nil {
		err = r.finish(manifestPath, &result, "failed", fmt.Errorf("encode worker request: %w", err))
		return result, err
	}
	if len(requestData) > MaxRequest {
		err = r.finish(manifestPath, &result, "failed", fmt.Errorf("worker request exceeds %d bytes", MaxRequest))
		return result, err
	}
	cmd := r.command(ctx)
	if cmd == nil {
		err = r.finish(manifestPath, &result, "failed", fmt.Errorf("worker command factory returned nil"))
		return result, err
	}
	cmd.Stdin = bytes.NewReader(requestData)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		err = r.finish(manifestPath, &result, "failed", err)
		return result, err
	}
	stderr := &limitedBuffer{limit: 64 << 10}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		err = r.finish(manifestPath, &result, "failed", err)
		return result, err
	}
	result.Manifest.Status = "running"
	result.Manifest.StartedAt = time.Now().UTC()
	if err := writeJSONAtomic(manifestPath, result.Manifest); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return result, err
	}

	scanErr := r.consume(stdout, directory, manifestPath, request.Plan.ID, &result)
	_ = stdout.Close()
	waitErr := cmd.Wait()
	terminalErr := errors.Join(scanErr, waitErr)
	if terminalErr == nil && result.Manifest.Status != "complete" {
		terminalErr = fmt.Errorf("worker exited without a complete final event")
	}
	if terminalErr != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			terminalErr = fmt.Errorf("%w: %s", terminalErr, detail)
		}
		status := "failed"
		if ctx.Err() != nil {
			status = "canceled"
			terminalErr = errors.Join(ctx.Err(), terminalErr)
		}
		terminalErr = r.finish(manifestPath, &result, status, terminalErr)
		return result, terminalErr
	}
	result.Manifest.FinishedAt = time.Now().UTC()
	if err := writeJSONAtomic(manifestPath, result.Manifest); err != nil {
		return result, err
	}
	return result, nil
}

func (r *Runner) consume(stdout io.Reader, directory, manifestPath, planID string, result *RunResult) error {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), MaxEvent)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("decode worker event: %w", err)
		}
		if event.Schema != EventSchema || event.PlanID != planID {
			return fmt.Errorf("invalid worker event identity")
		}
		switch event.Type {
		case "started":
			if event.Capture != nil {
				return fmt.Errorf("started event contains capture")
			}
		case "checkpoint", "final":
			if event.Capture == nil {
				return fmt.Errorf("%s event has no capture", event.Type)
			}
			if err := writeCaptureAtomic(filepath.Join(directory, "measurements.json"), *event.Capture); err != nil {
				return err
			}
			capture := *event.Capture
			result.Capture = &capture
			result.Manifest.Observations = len(capture.Run.Observations)
			if event.Type == "final" {
				if capture.Status == "complete" && event.Error == "" {
					result.Manifest.Status = "complete"
				} else {
					result.Manifest.Status = "failed"
					result.Manifest.Error = event.Error
					if result.Manifest.Error == "" {
						result.Manifest.Error = capture.Error
					}
				}
			}
			if err := writeJSONAtomic(manifestPath, result.Manifest); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown worker event type %q", event.Type)
		}
	}
	return scanner.Err()
}

func (r *Runner) finish(path string, result *RunResult, status string, err error) error {
	result.Manifest.Status = status
	result.Manifest.FinishedAt = time.Now().UTC()
	if err != nil {
		result.Manifest.Error = err.Error()
	}
	writeErr := writeJSONAtomic(path, result.Manifest)
	return errors.Join(err, writeErr)
}

// Recover marks non-terminal manifests left by a parent crash as failed while
// preserving any measurements checkpoint already present beside them.
func Recover(artifactRoot string) ([]Manifest, error) {
	entries, err := os.ReadDir(artifactRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var recovered []Manifest
	for _, entry := range entries {
		if !entry.IsDir() || !jobIDPattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(artifactRoot, entry.Name(), "manifest.json")
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		var manifest Manifest
		if json.Unmarshal(data, &manifest) != nil || manifest.Schema != manifestSchema || (manifest.Status != "queued" && manifest.Status != "running") {
			continue
		}
		manifest.Status = "failed"
		manifest.FinishedAt = time.Now().UTC()
		manifest.Error = "worker interrupted before final status"
		if err := writeJSONAtomic(path, manifest); err != nil {
			return recovered, err
		}
		recovered = append(recovered, manifest)
	}
	return recovered, nil
}

func writeCaptureAtomic(path string, capture evidence.Capture) error {
	var data bytes.Buffer
	if err := evidence.WriteCapture(&data, capture); err != nil {
		return err
	}
	if data.Len() > MaxArtifactBytes {
		return fmt.Errorf("measurement artifact exceeds %d bytes", MaxArtifactBytes)
	}
	return writeAtomic(path, data.Bytes())
}

type RetentionPolicy struct {
	MaxJobs int
	MaxAge  time.Duration
}

type retainedArtifact struct {
	id       string
	path     string
	finished time.Time
}

// Prune removes only terminal, well-formed job directories directly beneath
// artifactRoot. Running/queued jobs and unrecognized entries are never touched.
func Prune(artifactRoot string, policy RetentionPolicy, now time.Time) ([]string, error) {
	if artifactRoot == "" || policy.MaxJobs < 1 || policy.MaxAge <= 0 {
		return nil, fmt.Errorf("retention policy requires a root, positive job count, and positive age")
	}
	entries, err := os.ReadDir(artifactRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var artifacts []retainedArtifact
	for _, entry := range entries {
		if !entry.IsDir() || !jobIDPattern.MatchString(entry.Name()) {
			continue
		}
		path := filepath.Join(artifactRoot, entry.Name())
		data, readErr := os.ReadFile(filepath.Join(path, "manifest.json"))
		if readErr != nil {
			continue
		}
		var manifest Manifest
		if json.Unmarshal(data, &manifest) != nil || manifest.Schema != manifestSchema || manifest.JobID != entry.Name() ||
			(manifest.Status != "complete" && manifest.Status != "failed" && manifest.Status != "canceled") {
			continue
		}
		finished := manifest.FinishedAt
		if finished.IsZero() {
			finished = manifest.QueuedAt
		}
		artifacts = append(artifacts, retainedArtifact{id: entry.Name(), path: path, finished: finished})
	}
	slices.SortFunc(artifacts, func(a, b retainedArtifact) int { return b.finished.Compare(a.finished) })
	var removed []string
	for index, artifact := range artifacts {
		expired := !artifact.finished.IsZero() && now.Sub(artifact.finished) > policy.MaxAge
		if index < policy.MaxJobs && !expired {
			continue
		}
		relative, err := filepath.Rel(artifactRoot, artifact.path)
		if err != nil || relative != artifact.id || filepath.IsAbs(relative) {
			return removed, fmt.Errorf("artifact containment check failed for %q", artifact.id)
		}
		if err := os.RemoveAll(artifact.path); err != nil {
			return removed, err
		}
		removed = append(removed, artifact.id)
	}
	return removed, nil
}

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, byte(10))
	return writeAtomic(path, data)
}

func writeAtomic(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

type limitedBuffer struct {
	data  bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if b.data.Len() < b.limit {
		remaining := b.limit - b.data.Len()
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = b.data.Write(p)
	}
	return n, nil
}

func (b *limitedBuffer) String() string { return b.data.String() }
