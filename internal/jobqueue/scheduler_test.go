package jobqueue

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerSerializesJobs(t *testing.T) {
	s, err := New(Limits{MaxQueued: 4, MaxWallTime: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	start := make(chan struct{})
	release := make(chan struct{})
	var active, maximum atomic.Int32
	job := func(ctx context.Context) error {
		current := active.Add(1)
		for {
			previous := maximum.Load()
			if current <= previous || maximum.CompareAndSwap(previous, current) {
				break
			}
		}
		start <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			active.Add(-1)
			return ctx.Err()
		}
		active.Add(-1)
		return nil
	}
	results := make(chan Result, 2)
	go func() { results <- s.Run(context.Background(), job) }()
	<-start
	go func() { results <- s.Run(context.Background(), job) }()
	select {
	case <-start:
		t.Fatal("second job started concurrently")
	case <-time.After(20 * time.Millisecond):
	}
	release <- struct{}{}
	<-start
	release <- struct{}{}
	for range 2 {
		if result := <-results; result.Err != nil || result.StartedAt.IsZero() || result.FinishedAt.Before(result.StartedAt) {
			t.Fatalf("result = %+v", result)
		}
	}
	if maximum.Load() != 1 || s.Stats().Completed != 2 {
		t.Fatalf("maximum=%d stats=%+v", maximum.Load(), s.Stats())
	}
}

func TestSchedulerRejectsFullQueue(t *testing.T) {
	s, err := New(Limits{MaxQueued: 1, MaxWallTime: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	started := make(chan struct{})
	release := make(chan struct{})
	blocking := func(ctx context.Context) error {
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	first := make(chan Result, 1)
	go func() { first <- s.Run(context.Background(), blocking) }()
	<-started
	second := make(chan Result, 1)
	go func() { second <- s.Run(context.Background(), func(context.Context) error { return nil }) }()
	for deadline := time.Now().Add(time.Second); s.Stats().Queued != 1; {
		if time.Now().After(deadline) {
			t.Fatal("second job was not queued")
		}
		time.Sleep(time.Millisecond)
	}
	third := s.Run(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(third.Err, ErrQueueFull) || !third.StartedAt.IsZero() {
		t.Fatalf("third result = %+v", third)
	}
	close(release)
	if result := <-first; result.Err != nil {
		t.Fatal(result.Err)
	}
	if result := <-second; result.Err != nil {
		t.Fatal(result.Err)
	}
}

func TestSchedulerTimeoutAndQueuedCancellation(t *testing.T) {
	s, err := New(Limits{MaxQueued: 2, MaxWallTime: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	result := s.Run(context.Background(), func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() })
	if !errors.Is(result.Err, context.DeadlineExceeded) {
		t.Fatalf("timeout result = %+v", result)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	first := make(chan Result, 1)
	go func() {
		first <- s.Run(context.Background(), func(ctx context.Context) error {
			close(started)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	var ran atomic.Bool
	second := make(chan Result, 1)
	go func() { second <- s.Run(ctx, func(context.Context) error { ran.Store(true); return nil }) }()
	for deadline := time.Now().Add(time.Second); s.Stats().Queued != 1; {
		if time.Now().After(deadline) {
			t.Fatal("cancel target was not queued")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if queued := <-second; !errors.Is(queued.Err, context.Canceled) {
		t.Fatalf("queued result = %+v", queued)
	}
	close(release)
	if active := <-first; active.Err != nil {
		t.Fatal(active.Err)
	}
	for deadline := time.Now().Add(time.Second); s.Stats().Queued != 0; {
		if time.Now().After(deadline) {
			t.Fatal("cancelled request was not drained")
		}
		time.Sleep(time.Millisecond)
	}
	if ran.Load() {
		t.Fatal("cancelled queued job ran")
	}
}

func TestSchedulerCloseCancelsActiveAndRejectsNewWork(t *testing.T) {
	s, err := New(Limits{MaxQueued: 1, MaxWallTime: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	result := make(chan Result, 1)
	go func() {
		result <- s.Run(context.Background(), func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		})
	}()
	<-started
	done := make(chan struct{})
	go func() { s.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("close did not stop worker")
	}
	if r := <-result; !errors.Is(r.Err, ErrClosed) && !errors.Is(r.Err, context.Canceled) {
		t.Fatalf("active result = %+v", r)
	}
	if r := s.Run(context.Background(), func(context.Context) error { return nil }); !errors.Is(r.Err, ErrClosed) {
		t.Fatalf("post-close result = %+v", r)
	}
	s.Close()
}

func TestSchedulerValidation(t *testing.T) {
	for _, limits := range []Limits{{}, {MaxQueued: 1}, {MaxWallTime: time.Second}} {
		if _, err := New(limits); err == nil {
			t.Fatalf("accepted %+v", limits)
		}
	}
	s, err := New(Limits{MaxQueued: 1, MaxWallTime: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if r := s.Run(nil, func(context.Context) error { return nil }); r.Err == nil {
		t.Fatal("nil context accepted")
	}
	if r := s.Run(context.Background(), nil); r.Err == nil {
		t.Fatal("nil job accepted")
	}
}

func TestSchedulerConcurrentClose(t *testing.T) {
	s, err := New(Limits{MaxQueued: 1, MaxWallTime: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() { defer wg.Done(); s.Close() }()
	}
	wg.Wait()
}
