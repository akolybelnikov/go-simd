// Package jobqueue provides the bounded single-worker scheduler used to keep
// benchmark jobs from contaminating one another on a shared measuring process.
package jobqueue

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrClosed    = errors.New("job scheduler is closed")
	ErrQueueFull = errors.New("job queue is full")
)

type Limits struct {
	MaxQueued   int
	MaxWallTime time.Duration
}

func (l Limits) validate() error {
	if l.MaxQueued < 1 || l.MaxWallTime <= 0 {
		return fmt.Errorf("job queue limits must have positive queue and wall-time bounds")
	}
	return nil
}

// Job must observe ctx. The isolated-worker layer supplies a hard-stop fallback
// for a child process that does not exit after cancellation.
type Job func(ctx context.Context) error

type Result struct {
	QueuedAt   time.Time
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

type Stats struct {
	Queued    int
	Active    bool
	Completed uint64
}

type request struct {
	ctx      context.Context
	job      Job
	queuedAt time.Time
	result   chan Result
}

// Scheduler runs exactly one job at a time and admits at most MaxQueued waiting
// requests. Run is synchronous so cancellation and errors cannot be ignored.
type Scheduler struct {
	limits Limits
	queue  chan request
	root   context.Context
	cancel context.CancelFunc
	closed atomic.Bool
	active atomic.Bool
	done   atomic.Uint64
	wg     sync.WaitGroup
}

func New(limits Limits) (*Scheduler, error) {
	if err := limits.validate(); err != nil {
		return nil, err
	}
	root, cancel := context.WithCancel(context.Background())
	s := &Scheduler{limits: limits, queue: make(chan request, limits.MaxQueued), root: root, cancel: cancel}
	s.wg.Add(1)
	go s.loop()
	return s, nil
}

// Run admits a job without waiting for queue capacity. A full queue is a stable
// error rather than unbounded backpressure. Waiting callers may cancel safely;
// canceled queued jobs are skipped when they reach the worker.
func (s *Scheduler) Run(ctx context.Context, job Job) Result {
	queuedAt := time.Now().UTC()
	if ctx == nil {
		return Result{QueuedAt: queuedAt, FinishedAt: queuedAt, Err: fmt.Errorf("nil job context")}
	}
	if job == nil {
		return Result{QueuedAt: queuedAt, FinishedAt: queuedAt, Err: fmt.Errorf("nil job")}
	}
	if err := ctx.Err(); err != nil {
		return Result{QueuedAt: queuedAt, FinishedAt: queuedAt, Err: err}
	}
	if s.closed.Load() {
		return Result{QueuedAt: queuedAt, FinishedAt: queuedAt, Err: ErrClosed}
	}
	req := request{ctx: ctx, job: job, queuedAt: queuedAt, result: make(chan Result, 1)}
	select {
	case <-s.root.Done():
		return Result{QueuedAt: queuedAt, FinishedAt: time.Now().UTC(), Err: ErrClosed}
	case s.queue <- req:
	default:
		return Result{QueuedAt: queuedAt, FinishedAt: time.Now().UTC(), Err: ErrQueueFull}
	}
	select {
	case result := <-req.result:
		return result
	case <-ctx.Done():
		return Result{QueuedAt: queuedAt, FinishedAt: time.Now().UTC(), Err: ctx.Err()}
	case <-s.root.Done():
		return Result{QueuedAt: queuedAt, FinishedAt: time.Now().UTC(), Err: ErrClosed}
	}
}

func (s *Scheduler) Stats() Stats {
	return Stats{Queued: len(s.queue), Active: s.active.Load(), Completed: s.done.Load()}
}

// Close cancels the active job, rejects new work and waits for the scheduler
// goroutine. It is safe to call more than once.
func (s *Scheduler) Close() {
	if s.closed.CompareAndSwap(false, true) {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()
	for {
		select {
		case <-s.root.Done():
			s.failQueued(ErrClosed)
			return
		case req := <-s.queue:
			s.execute(req)
		}
	}
}

func (s *Scheduler) execute(req request) {
	if err := req.ctx.Err(); err != nil {
		s.deliver(req, Result{QueuedAt: req.queuedAt, FinishedAt: time.Now().UTC(), Err: err})
		return
	}
	startedAt := time.Now().UTC()
	ctx, cancel := context.WithTimeout(s.root, s.limits.MaxWallTime)
	stopRequest := context.AfterFunc(req.ctx, cancel)
	s.active.Store(true)
	err := req.job(ctx)
	s.active.Store(false)
	stopRequest()
	contextErr := ctx.Err()
	cancel()
	if err == nil && contextErr != nil {
		err = contextErr
	}
	s.done.Add(1)
	s.deliver(req, Result{QueuedAt: req.queuedAt, StartedAt: startedAt, FinishedAt: time.Now().UTC(), Err: err})
}

func (s *Scheduler) deliver(req request, result Result) {
	select {
	case req.result <- result:
	default:
	}
}

func (s *Scheduler) failQueued(err error) {
	for {
		select {
		case req := <-s.queue:
			s.deliver(req, Result{QueuedAt: req.queuedAt, FinishedAt: time.Now().UTC(), Err: err})
		default:
			return
		}
	}
}
