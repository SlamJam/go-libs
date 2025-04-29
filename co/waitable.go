package co

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type Waitable interface {
	Wait(ctx context.Context) error
}

func All(ctx context.Context, ws ...Waitable) error {
	var result error

	for _, w := range ws {
		if multierr.AppendInto(&result, w.Wait(ctx)) {
			break
		}
	}

	return result
}

// func All(ws ...Waitable) error {
// 	return AllCtx(context.Background(), ws...)
// }

// func AllTimeout(d time.Duration, ws ...Waitable) error {
// 	ctx, cancel := context.WithTimeoutCause(context.Background(), d, std.ErrTimeout)
// 	defer cancel()

// 	return AllCtx(ctx, ws...)
// }

func FirstError(ctx context.Context, ws ...Waitable) (result error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)

	var wg sync.WaitGroup
	for _, w := range ws {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- w.Wait(ctx)
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
