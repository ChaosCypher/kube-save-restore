package workerpool

import (
	"context"
	"errors"
	"sync"
)

// Task represents a function that can be executed by the worker pool.
type Task func(ctx context.Context) error

// WorkerPool manages a pool of workers to execute tasks concurrently.
type WorkerPool struct {
	tasks          chan Task
	wg             sync.WaitGroup
	maxConcurrency int
	errors         []error
	mu             sync.Mutex
	closed         bool
	muClose        sync.RWMutex
}

// NewWorkerPool creates a new WorkerPool with the specified maximum concurrency.
func NewWorkerPool(maxConcurrency, taskQueueSize int) *WorkerPool {
	return &WorkerPool{
		tasks:          make(chan Task, taskQueueSize),
		maxConcurrency: maxConcurrency,
	}
}

// AddTask adds a new task to the worker pool.
// It returns an error if the pool is already closed.
func (wp *WorkerPool) AddTask(task Task) error {
	wp.muClose.RLock()
	if wp.closed {
		wp.muClose.RUnlock()
		return errors.New("worker pool is closed")
	}
	wp.muClose.RUnlock()

	select {
	case wp.tasks <- task:
		return nil
	default:
		return errors.New("task queue is full")
	}
}

// Close closes the worker pool, preventing new tasks from being added.
func (wp *WorkerPool) Close() {
	wp.muClose.Lock()
	defer wp.muClose.Unlock()
	if !wp.closed {
		close(wp.tasks)
		wp.closed = true
	}
}

// Run starts the worker pool and executes all tasks.
// It returns a slice of errors encountered during task execution.
func (wp *WorkerPool) Run(ctx context.Context) []error {
	for i := 0; i < wp.maxConcurrency; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}

	wp.wg.Wait()

	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.errors
}

// worker is the main worker function that processes tasks from the channel.
func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			if err := task(ctx); err != nil {
				wp.mu.Lock()
				wp.errors = append(wp.errors, err)
				wp.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
