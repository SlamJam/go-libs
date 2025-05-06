package co

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPromiseResolution(t *testing.T) {
	// Test a promise that resolves successfully
	p := NewResolved(42)

	// Check that the promise is completed
	if !p.IsCompleted() {
		t.Error("Expected promise to be completed")
	}

	// Check the resolved value
	result, err := p.Value()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 42 {
		t.Errorf("Expected value 42, got %v", result)
	}

	// Test Await
	err = p.Await(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test Poll
	val, err := p.Poll(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %v", val)
	}
}

func TestPromiseRejection(t *testing.T) {
	// Test a promise that rejects with an error
	expectedErr := errors.New("test error")
	p := NewRejected[int](expectedErr)

	// Check that the promise is completed
	if !p.IsCompleted() {
		t.Error("Expected promise to be completed")
	}

	// Test Await
	err := p.Await(context.Background())
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Test Poll
	_, err = p.Poll(context.Background())
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Test Value
	_, err = p.Value()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestLazyPromise(t *testing.T) {
	executed := false

	p := NewLazyPromise(func() (int, error) {
		executed = true
		return 10, nil
	})

	// Lazy promise should not be launched immediately
	if p.IsLaunched() {
		t.Error("Expected lazy promise not to be launched")
	}

	if executed {
		t.Error("Lazy promise function should not be executed until polled")
	}

	// Poll to trigger execution
	result, err := p.Poll(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 10 {
		t.Errorf("Expected result 10, got %v", result)
	}

	// Now it should be launched and executed
	if !p.IsLaunched() {
		t.Error("Expected promise to be launched after Poll")
	}
	if !executed {
		t.Error("Expected function to be executed after Poll")
	}
}

func TestEagerPromise(t *testing.T) {
	var executed atomic.Bool

	p := NewPromise(func() (int, error) {
		executed.Store(true)
		return 20, nil
	})

	// Eager promise should be launched immediately
	if !p.IsLaunched() {
		t.Error("Expected eager promise to be launched immediately")
	}

	// Wait a bit to ensure the function has time to execute
	time.Sleep(10 * time.Millisecond)

	if !executed.Load() {
		t.Error("Eager promise function should be executed immediately")
	}

	// Check the result
	result, err := p.Poll(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 20 {
		t.Errorf("Expected result 20, got %v", result)
	}
}

func TestPromiseWithTimeout(t *testing.T) {
	// Create a promise that takes time to complete
	p := NewPromise(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 30, nil
	})

	// Create a context with timeout shorter than the promise execution
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Polling with a timeout should return context error
	_, err := p.Poll(ctx)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected deadline exceeded error, got %v", err)
	}

	// The promise should still complete eventually
	time.Sleep(200 * time.Millisecond)
	if !p.IsCompleted() {
		t.Error("Expected promise to complete despite timeout")
	}

	// Now we should be able to get the result
	result, err := p.Poll(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != 30 {
		t.Errorf("Expected result 30, got %v", result)
	}
}

func TestPanicOnUninitializedPromise(t *testing.T) {
	// Create an uninitialized promise
	p := &promise[int]{}

	// Calling methods on uninitialized promise should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on uninitialized promise")
		}
	}()

	p.Poll(context.Background())
}

func TestUncompletedValuePanic(t *testing.T) {
	// Create a promise that takes time to complete
	p := NewLazyPromise(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 40, nil
	})

	// Value() should panic if called before the promise is completed
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Value() to panic for uncompleted promise")
		}
	}()

	p.Value() // This should panic
}
