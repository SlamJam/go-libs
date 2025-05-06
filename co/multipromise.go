package co

import (
	"context"
	"iter"
	"sync"

	std "github.com/SlamJam/go-libs"
	"github.com/SlamJam/go-libs/xchan"
	"github.com/SlamJam/go-libs/xslices"
	"go.uber.org/multierr"
)

type MultiPromise[T any] []Promise[T]

func (mp *MultiPromise[T]) Add(f func() (T, error)) {
	*mp = append(*mp, NewPromise(f))
}

func (mp *MultiPromise[T]) Append(p Promise[T]) {
	*mp = append(*mp, p)
}

func (mp MultiPromise[T]) AsWaitables() []Awaitable {
	return xslices.Map(mp, func(p Promise[T]) Awaitable { return p })
}

// AllResults дожидается выполнения всех задач
// Возвращает или все результаты или все ошибки, собранные в multierr
func (mp MultiPromise[T]) AllResults(ctx context.Context) ([]T, error) {
	results := make([]T, 0, len(mp))
	var err error
	for _, item := range mp.IterAllResults(ctx) {
		if !multierr.AppendInto(&err, item.Err) && err == nil {
			results = append(results, item.Result)
		} else {
			results = nil
		}
	}

	return results, err
}

// AllResultsOrFirstError дожидается выполнения всех задач или первой возникшей ошибки
// Возвращает или все результаты или первую возникшую ошибку
func (mp MultiPromise[T]) AllResultsOrFirstError(ctx context.Context) ([]T, error) {
	results := make([]T, 0, len(mp))
	for _, item := range mp.IterAllResults(ctx) {
		if err := item.Err; err != nil {
			return nil, err
		}

		results = append(results, item.Result)
	}

	return results, nil
}

func (mp MultiPromise[T]) FirstResult(ctx context.Context) (T, error) {
	var err error
	for _, item := range mp.IterAllResults(ctx) {
		if !multierr.AppendInto(&err, item.Err) {
			return item.Result, nil
		}
	}

	return std.Zero[T](), err
}

type MultiPromiseIterResult[T any] struct {
	Result T
	Err    error
}

type iterIndexedResult[T any] struct {
	Index int
	Value MultiPromiseIterResult[T]
}

func (mp MultiPromise[T]) iterResults(
	ctx context.Context,
	iterFunc func(chan iterIndexedResult[T]) iter.Seq2[int, MultiPromiseIterResult[T]],
) iter.Seq2[int, MultiPromiseIterResult[T]] {

	ch := make(chan iterIndexedResult[T])
	// Если мы вышли раньше и не дочитали до конца, нужно дрейнить каналы, пока в них кто-то пишет
	defer xchan.Drain(ch)

	var wg sync.WaitGroup
	for idx, p := range mp {
		wg.Add(1)
		go func() {
			defer wg.Done()

			res, err := p.Poll(ctx)
			ch <- iterIndexedResult[T]{Index: idx, Value: MultiPromiseIterResult[T]{Result: res, Err: err}}
		}()
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	return iterFunc(ch)
}

func (mp MultiPromise[T]) IterAllResults(ctx context.Context) iter.Seq2[int, MultiPromiseIterResult[T]] {
	iterFunc := func(ch chan iterIndexedResult[T]) iter.Seq2[int, MultiPromiseIterResult[T]] {
		return func(yield func(int, MultiPromiseIterResult[T]) bool) { // iter.Seq2[int, MultiPromiseIterResult[T]]
			for item := range ch {
				if !yield(item.Index, item.Value) {
					return
				}
			}
		}
	}

	return mp.iterResults(ctx, iterFunc)
}

func (mp MultiPromise[T]) IterResultsUntillCancel(ctx context.Context) iter.Seq2[int, MultiPromiseIterResult[T]] {
	iterFunc := func(ch chan iterIndexedResult[T]) iter.Seq2[int, MultiPromiseIterResult[T]] {
		return func(yield func(int, MultiPromiseIterResult[T]) bool) { // iter.Seq2[int, MultiPromiseIterResult[T]]
			for {
				select {
				case item, ok := <-ch:
					if !ok {
						return
					}
					if !yield(item.Index, item.Value) {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}

	return mp.iterResults(ctx, iterFunc)
}

type IndexedResult[T any] struct {
	Index int
	Value T
}

type IndexedError struct {
	Index int
	Err   error
}

type MultiPromisePartialResult[T any] struct {
	Results []IndexedResult[T]
	Errors  []IndexedError
}

func (pr *MultiPromisePartialResult[T]) addError(idx int, err error) {
	pr.Errors = append(pr.Errors, IndexedError{Index: idx, Err: err})
}

func (pr *MultiPromisePartialResult[T]) addResult(idx int, val T) {
	pr.Results = append(pr.Results, IndexedResult[T]{Index: idx, Value: val})
}

func (pr *MultiPromisePartialResult[T]) AvailableResults() []T {
	result := make([]T, 0, len(pr.Results))
	for _, r := range pr.Results {
		result = append(result, r.Value)
	}

	return result
}

func (pr *MultiPromisePartialResult[T]) MultiErr() error {
	var result error
	for _, err := range pr.Errors {
		result = multierr.Append(result, err.Err)
	}

	return result
}

func (mp MultiPromise[T]) PartialResult(ctx context.Context) (result MultiPromisePartialResult[T]) {
	for idx, item := range mp.IterAllResults(ctx) {
		if item.Err != nil {
			result.addError(idx, item.Err)
		} else {
			result.addResult(idx, item.Result)
		}
	}

	return
}
