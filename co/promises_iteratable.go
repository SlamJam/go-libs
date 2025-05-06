package co

import (
	"context"
	"sync"

	"github.com/SlamJam/go-libs/xchan"
)

type PromiseWithKey[T, K any] struct {
	Key K
	Promise[T]
}

type IterResultItem[T any] struct {
	Result T
	Err    error
}

func IterAllResults[T, K any](ctx context.Context, promises ...PromiseWithKey[T, K]) Iterator[T, K] {
	return func(yield func(K, IterResultItem[T]) bool) {
		ch := InternalKeyedResultChan(ctx, promises...)
		defer xchan.Drain(ch)

		for item := range ch {
			if !yield(item.Key, item.Value) {
				return
			}
		}
	}
}

func IterResultsUntilCancel[T, K any](ctx context.Context, promises ...PromiseWithKey[T, K]) Iterator[T, K] {
	return func(yield func(K, IterResultItem[T]) bool) {
		ch := InternalKeyedResultChan(ctx, promises...)
		defer xchan.Drain(ch)

		for {
			select {
			case item, ok := <-ch:
				if !ok {
					return
				}
				if !yield(item.Key, item.Value) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

type iterResultWithKey[T, K any] struct {
	Key   K
	Value IterResultItem[T]
}

// Result channel MUST be drained
func InternalKeyedResultChan[T, K any](
	ctx context.Context,
	promises ...PromiseWithKey[T, K],
) <-chan iterResultWithKey[T, K] {
	ch := make(chan iterResultWithKey[T, K])

	var wg sync.WaitGroup
	for _, p := range promises {
		wg.Add(1)
		go func() {
			defer wg.Done()

			res, err := p.Poll(ctx)
			ch <- iterResultWithKey[T, K]{Key: p.Key, Value: IterResultItem[T]{Result: res, Err: err}}
		}()
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	return ch
}
