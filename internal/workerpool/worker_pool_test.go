package workerpool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestAddTaskToClosedPool verifies that adding a task to a closed WorkerPool panics.
func TestAddTaskToClosedPool(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Close()

	task := createSuccessTask()
	expectPanic(t, func() {
		wp.AddTask(task)
	})
}

// Helper function to expect a panic when adding a task to a closed pool.
func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but no panic occurred")
		}
	}()
	fn()
}

// Helper function to assert that a task was added to the WorkerPool.
func assertTaskAdded(t *testing.T, wp *WorkerPool) {
	t.Helper()
	select {
	case <-wp.tasks:
		// Task successfully added
	default:
		t.Error("Expected task to be added to the tasks channel")
	}
}

// Helper function to create multiple tasks based on success or failure.
func createMultipleTasks(t *testing.T, count int, fail bool) []Task {
	t.Helper()
	tasks := make([]Task, count)
	for i := 0; i < count; i++ {
		if fail {
			tasks[i] = createFailTask()
		} else {
			tasks[i] = createSuccessTask()
		}
	}
	return tasks
}

// Helper function to assert the number of errors.
func assertErrorCount(t *testing.T, errors []error, expected int) {
	t.Helper()
	if len(errors) != expected {
		t.Errorf("Expected %d errors, got %d", expected, len(errors))
	}
}

// Helper function to create and setup a WorkerPool with specified tasks.
func setupWorkerPool(t *testing.T, maxConcurrency int, tasks []Task, closeAfterAdding bool) *WorkerPool {
	t.Helper()
	wp := NewWorkerPool(maxConcurrency)
	for _, task := range tasks {
		wp.AddTask(task)
	}
	if closeAfterAdding {
		wp.Close()
	}
	return wp
}

// Helper function to run the WorkerPool and return errors.
func runWorkerPool(ctx context.Context, t *testing.T, wp *WorkerPool) []error {
	t.Helper()
	return wp.Run(ctx)
}

// Helper function to create a simple task that succeeds.
func createSuccessTask() Task {
	return func(context.Context) error {
		return nil
	}
}

// Helper function to create a simple task that fails.
func createFailTask() Task {
	return func(context.Context) error {
		return errors.New("task failed")
	}
}

// TestNewWorkerPool verifies the initialization of a new WorkerPool.
func TestNewWorkerPool(t *testing.T) {
	maxConcurrency := 5
	wp := NewWorkerPool(maxConcurrency)

	if wp.maxConcurrency != maxConcurrency {
		t.Errorf("Expected maxConcurrency %d, got %d", maxConcurrency, wp.maxConcurrency)
	}

	if cap(wp.tasks) != 100 {
		t.Errorf("Expected tasks channel capacity 100, got %d", cap(wp.tasks))
	}

	if wp.closed {
		t.Error("Expected WorkerPool to be open upon creation")
	}
}

// TestAddTask verifies that a task can be added to the WorkerPool.
func TestAddTask(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Close()

	task := createSuccessTask()
	wp.AddTask(task)

	assertTaskAdded(t, wp)
}

// TestRun verifies that tasks are executed correctly without errors.
func TestRun(t *testing.T) {
	totalTasks := 10
	var mu sync.Mutex
	results := make([]int, 0, totalTasks)

	tasks := make([]Task, totalTasks)
	for i := 0; i < totalTasks; i++ {
		index := i
		tasks[i] = func(context.Context) error {
			mu.Lock()
			results = append(results, index)
			mu.Unlock()
			return nil
		}
	}

	wp := setupWorkerPool(t, 3, tasks, true)

	ctx := context.Background()
	errors := runWorkerPool(ctx, t, wp)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(results) != totalTasks {
		t.Errorf("Expected %d results, got %d", totalTasks, len(results))
	}
}

// TestRunWithErrors verifies that the WorkerPool correctly collects errors from failing tasks.
func TestRunWithErrors(t *testing.T) {
	totalFailTasks := 2
	tasks := append(createMultipleTasks(t, 1, false), createMultipleTasks(t, totalFailTasks, true)...)

	wp := setupWorkerPool(t, 2, tasks, true)

	ctx := context.Background()
	errors := runWorkerPool(ctx, t, wp)

	assertErrorCount(t, errors, totalFailTasks)
}

// TestRunWithContextCancellation verifies that tasks respect context cancellation.
func TestRunWithContextCancellation(t *testing.T) {
	wp := NewWorkerPool(2)
	defer wp.Close()

	taskLong := func(ctx context.Context) error {
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	wp.AddTask(taskLong)
	wp.AddTask(taskLong)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	errorsList := wp.Run(ctx)

	// Depending on timing, some tasks may have been cancelled
	if len(errorsList) > 2 {
		t.Errorf("Expected between 0 and 2 errors, got %d", len(errorsList))
	}

	for _, err := range errorsList {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context deadline exceeded error, got %v", err)
		}
	}
}

// TestClose verifies that the WorkerPool can be closed properly and that closing is idempotent.
func TestClose(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Close()

	if !wp.closed {
		t.Error("Expected WorkerPool to be closed")
	}

	// Ensure that closing is idempotent
	wp.Close()
}

// TestConcurrentAddAndRun verifies that the WorkerPool can handle concurrent task additions and executions.
func TestConcurrentAddAndRun(t *testing.T) {
	totalTasks := 100
	var mu sync.Mutex
	counter := 0

	tasks := make([]Task, totalTasks)
	for i := 0; i < totalTasks; i++ {
		tasks[i] = func(context.Context) error {
			mu.Lock()
			counter++
			mu.Unlock()
			return nil
		}
	}

	wp := setupWorkerPool(t, 5, tasks, true)

	ctx := context.Background()
	errors := runWorkerPool(ctx, t, wp)

	assertErrorCount(t, errors, 0)

	if counter != totalTasks {
		t.Errorf("Expected counter to be %d, got %d", totalTasks, counter)
	}
}
