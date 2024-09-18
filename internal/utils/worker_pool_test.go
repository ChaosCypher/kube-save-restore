package utils

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestNewWorkerPool verifies the correct initialization of a WorkerPool.
func TestNewWorkerPool(t *testing.T) {
	maxConcurrency := 5
	wp := NewWorkerPool(maxConcurrency)

	if wp == nil {
		t.Fatal("Expected WorkerPool instance, got nil")
	}

	if wp.maxConcurrency != maxConcurrency {
		t.Errorf("Expected maxConcurrency to be %d, got %d", maxConcurrency, wp.maxConcurrency)
	}

	if wp.tasks == nil {
		t.Error("Expected tasks channel to be initialized, got nil")
	}

	if len(wp.errors) != 0 {
		t.Error("Expected errors slice to be empty")
	}
}

// TestAddTask verifies that tasks are correctly added to the WorkerPool.
func TestAddTask(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Close() // Ensure the channel is closed after the test

	var received Task
	var done bool

	// Start a goroutine to receive the task
	go func() {
		received = <-wp.tasks
		done = true
	}()

	sampleTask := func(ctx context.Context) error {
		return nil
	}

	wp.AddTask(sampleTask)

	// Wait briefly to allow the goroutine to receive the task
	time.Sleep(100 * time.Millisecond)

	if !done {
		t.Error("Expected a task to be added and received, but it wasn't")
	}

	if received == nil {
		t.Error("Expected a task to be added, but got nil")
	}
}

// TestRun tests the Run method of WorkerPool with various scenarios.
func TestRun(t *testing.T) {
	// Test case: No tasks added to the pool
	t.Run("NoTasks", func(t *testing.T) {
		wp := NewWorkerPool(3)
		ctx := context.Background()

		// Close immediately since no tasks are added
		wp.Close()

		errors := wp.Run(ctx)
		if len(errors) != 0 {
			t.Errorf("Expected no errors, got %d", len(errors))
		}
	})

	// Test case: All tasks succeed
	t.Run("AllTasksSucceed", func(t *testing.T) {
		wp := NewWorkerPool(3)
		ctx := context.Background()

		numTasks := 10
		for i := 0; i < numTasks; i++ {
			wp.AddTask(func(ctx context.Context) error {
				return nil
			})
		}

		// Close after adding all tasks to signal completion
		wp.Close()

		errors := wp.Run(ctx)
		if len(errors) != 0 {
			t.Errorf("Expected no errors, got %d", len(errors))
		}
	})

	// Test case: Some tasks fail
	t.Run("SomeTasksFail", func(t *testing.T) {
		wp := NewWorkerPool(3)
		ctx := context.Background()

		numTasks := 10
		numFailures := 3
		var expectedErrors int

		for i := 0; i < numTasks; i++ {
			if i < numFailures {
				wp.AddTask(func(ctx context.Context) error {
					return errors.New("task failed")
				})
				expectedErrors++
			} else {
				wp.AddTask(func(ctx context.Context) error {
					return nil
				})
			}
		}

		// Close after adding all tasks
		wp.Close()

		errors := wp.Run(ctx)
		if len(errors) != expectedErrors {
			t.Errorf("Expected %d errors, got %d", expectedErrors, len(errors))
		}
	})

	// Test case: Context is cancelled during task execution
	t.Run("ContextCancelled", func(t *testing.T) {
		wp := NewWorkerPool(3)
		defer wp.Close()

		// Create a cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		numTasks := 5
		var wg sync.WaitGroup
		wg.Add(numTasks)

		// Add tasks that respect context cancellation
		for i := 0; i < numTasks; i++ {
			wp.AddTask(func(ctx context.Context) error {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return nil
				}
			})
		}

		// Channel to receive errors from Run
		errorsCh := make(chan []error, 1)

		// Start Run in a separate goroutine
		go func() {
			errors := wp.Run(ctx)
			errorsCh <- errors
		}()

		// Wait a short moment to allow tasks to start
		time.Sleep(50 * time.Millisecond)

		// Cancel the context while tasks are running
		cancel()

		// Wait for all tasks to acknowledge cancellation
		wg.Wait()

		// Retrieve errors from Run
		select {
		case errors := <-errorsCh:
			if len(errors) != numTasks {
				t.Errorf("Expected %d errors due to context cancellation, got %d", numTasks, len(errors))
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("Run method did not return within expected time")
		}
	})
}
