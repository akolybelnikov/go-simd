package labservice

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/akolybelnikov/go-simd/internal/jobqueue"
	"github.com/akolybelnikov/go-simd/internal/labworker"
)

type blockingWorker struct {
	active    atomic.Int32
	maxActive atomic.Int32
	release   <-chan struct{}
}

func (w *blockingWorker) Run(ctx context.Context, _ string, request labworker.Request) (labworker.RunResult, error) {
	active := w.active.Add(1)
	defer w.active.Add(-1)
	for {
		maximum := w.maxActive.Load()
		if active <= maximum || w.maxActive.CompareAndSwap(maximum, active) {
			break
		}
	}
	select {
	case <-w.release:
		return labworker.RunResult{Manifest: labworker.Manifest{JobID: request.JobID, Status: "complete"}}, nil
	case <-ctx.Done():
		return labworker.RunResult{}, ctx.Err()
	}
}

func TestExecutorSerializesWorkerProcesses(t *testing.T) {
	release := make(chan struct{})
	worker := &blockingWorker{release: release}
	executor, err := NewExecutor(t.TempDir(), jobqueue.Limits{MaxQueued: 2, MaxWallTime: time.Second}, worker)
	if err != nil {
		t.Fatal(err)
	}
	defer executor.Close()
	var wg sync.WaitGroup
	for _, id := range []string{"one", "two"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = executor.Execute(context.Background(), labworker.Request{JobID: id})
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	wg.Wait()
	if worker.maxActive.Load() != 1 {
		t.Fatalf("maximum active workers = %d", worker.maxActive.Load())
	}
}
