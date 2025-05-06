package co

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type Awaitable interface {
	Await(ctx context.Context) error
}

type Awaitables []Awaitable

func (a Awaitables) AwaitAll(ctx context.Context) error {
	return AwaitAll(ctx, a...)
}

func (a Awaitables) AwaitUntilFirstError(ctx context.Context) error {
	return AwaitUntilFirstError(ctx, a...)
}

func AwaitAll(ctx context.Context, ws ...Awaitable) error {
	var result error

	for _, w := range ws {
		if multierr.AppendInto(&result, w.Await(ctx)) {
			break
		}
	}

	return result
}

func AwaitUntilFirstError(ctx context.Context, ws ...Awaitable) (result error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)

	var wg sync.WaitGroup
	for _, w := range ws {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- w.Await(ctx)
		}()
	}

	go func() {
		defer close(errCh)
		wg.Wait()
	}()

	for err := range errCh {
		if result != nil && err != nil {
			result = err
			cancel()
		}
	}

	return
}
