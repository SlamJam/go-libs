package co

import (
	"context"
	"iter"
	"sync"

	std "github.com/SlamJam/go-libs"
	"github.com/SlamJam/go-libs/pair"
	"github.com/SlamJam/go-libs/xchan"
	"github.com/SlamJam/go-libs/xslices"
	"go.uber.org/multierr"
)

type MultiPromise[T any] []Promise[T]

func (mp *MultiPromise[T]) Add(f Deffered[T]) {
	*mp = append(*mp, NewPromise(f))
}

func (mp MultiPromise[T]) AsWaitables() []Waitable {
	return xslices.Map(mp, func(p Promise[T]) Waitable { return p })
}

func (mp MultiPromise[T]) results() []T {
	return xslices.Map(mp, func(p Promise[T]) T { return p.Value() })
}

// AllResults дожидается выполнения всех задач
// Возвращает или все результаты или все ошибки, собранные в multierr
func (mp MultiPromise[T]) AllResults(ctx context.Context) ([]T, error) {
	ws := mp.AsWaitables()

	err := All(ctx, ws...)
	if err != nil {
		return nil, err
	}

	return mp.results(), nil
}

// AllResultsOrFirstError дожидается выполнения всех задач или первой возникшей ошибки
// Возвращает или все результаты или первую возникшую ошибку
func (mp MultiPromise[T]) AllResultsOrFirstError(ctx context.Context) ([]T, error) {
	ws := mp.AsWaitables()

	err := FirstError(ctx, ws...)
	if err != nil {
		return nil, err
	}

	return mp.results(), nil
}

func (mp MultiPromise[T]) FirstResult(ctx context.Context) (T, error) {
	var err error
	for _, item := range mp.IterResultCtx(ctx) {
		if !multierr.AppendInto(&err, item.Second) {
			return item.First, nil
		}
	}

	return std.Zero[T](), err
}

func (mp MultiPromise[T]) IterResultCtx(ctx context.Context) iter.Seq2[int, pair.Pair[T, error]] {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type RES struct {
		Index      int
		ResultPair pair.Pair[T, error]
	}

	ch := make(chan RES)
	// Если мы вышли раньше и не дочитали до конца, нужно дрейнить каналы, пока в них кто-то пишет
	defer xchan.Drain(ch)

	var wg sync.WaitGroup
	for idx, p := range mp {
		wg.Add(1)
		go func() {
			defer wg.Done()

			res, err := p.Poll(ctx)
			ch <- RES{Index: idx, ResultPair: pair.New(res, err)}
		}()
	}

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	return func(yield func(int, pair.Pair[T, error]) bool) {
		// for item := range ch {
		// 	if !yield(item.Index, item.ResultPair) {
		// 		return
		// 	}
		// }

		for {
			select {
			case item, ok := <-ch:
				if !ok {
					return
				}
				if !yield(item.Index, item.ResultPair) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
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

func (mp MultiPromise[T]) PartialResultCtx(ctx context.Context) (result MultiPromisePartialResult[T]) {
	res, err := mp.AllResults(ctx)
	if err == nil {
		for idx, item := range res {
			result.Results = append(result.Results, IndexedResult[T]{Index: idx, Value: item})
		}

		return
	}

	for idx, p := range mp {
		if !p.IsCompleted() {
			continue
		}

		res, err := p.Poll(ctx)
		if err == nil {
			result.Results = append(result.Results, IndexedResult[T]{Index: idx, Value: res})
		} else {
			result.Errors = append(result.Errors, IndexedError{Index: idx, Err: err})
		}
	}

	return
}
