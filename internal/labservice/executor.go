// Package labservice composes bounded scheduling with isolated laboratory
// workers. Frontends and future typed tools depend on this layer.
package labservice

import (
	"context"
	"fmt"

	"github.com/akolybelnikov/go-simd/internal/jobqueue"
	"github.com/akolybelnikov/go-simd/internal/labworker"
)

type Worker interface {
	Run(context.Context, string, labworker.Request) (labworker.RunResult, error)
}

type Executor struct {
	scheduler *jobqueue.Scheduler
	worker    Worker
	root      string
}

func NewExecutor(root string, limits jobqueue.Limits, worker Worker) (*Executor, error) {
	if root == "" {
		return nil, fmt.Errorf("artifact root is required")
	}
	if worker == nil {
		return nil, fmt.Errorf("worker is required")
	}
	scheduler, err := jobqueue.New(limits)
	if err != nil {
		return nil, err
	}
	return &Executor{scheduler: scheduler, worker: worker, root: root}, nil
}

// Execute admits the request without unbounded waiting and serializes all
// measurements owned by this service instance.
func (e *Executor) Execute(ctx context.Context, request labworker.Request) (labworker.RunResult, error) {
	var output labworker.RunResult
	jobResult := e.scheduler.Run(ctx, func(jobCtx context.Context) error {
		var err error
		output, err = e.worker.Run(jobCtx, e.root, request)
		return err
	})
	return output, jobResult.Err
}

func (e *Executor) Stats() jobqueue.Stats { return e.scheduler.Stats() }

func (e *Executor) Close() { e.scheduler.Close() }
