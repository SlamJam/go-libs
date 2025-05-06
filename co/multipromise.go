package co

import (
	"context"
	"sort"

	std "github.com/SlamJam/go-libs"
	"github.com/SlamJam/go-libs/xslices"
	"github.com/pkg/errors"
)

var ErrEmptyMultiPromise = errors.New("empty MultiPromise")

type MultiPromise[T any] []Promise[T]

func (mp *MultiPromise[T]) Add(f func() (T, error)) {
	mp.Append(NewPromise(f))
}

func (mp *MultiPromise[T]) Append(p Promise[T]) {
	*mp = append(*mp, p)
}

func (mp MultiPromise[T]) AsWaitables() []Awaitable {
	return xslices.Map(mp, func(p Promise[T]) Awaitable { return p })
}

func (mp MultiPromise[T]) AsPromiseWithKeys() []PromiseWithKey[T, int] {
	result := make([]PromiseWithKey[T, int], 0, len(mp))
	for i, p := range mp {
		result = append(result, PromiseWithKey[T, int]{Key: i, Promise: p})
	}

	return result
}

func (mp MultiPromise[T]) iterAllResults(ctx context.Context) Iterator[T, int] {
	return IterAllResults(ctx, mp.AsPromiseWithKeys()...)
}

func (mp MultiPromise[T]) iterResultsUntillCancel(ctx context.Context) Iterator[T, int] {
	return IterResultsUntilCancel(ctx, mp.AsPromiseWithKeys()...)
}

// AllResults дожидается выполнения всех задач
// Возвращает или все результаты или все ошибки, собранные в multierr
func (mp MultiPromise[T]) AllResults(ctx context.Context) ([]T, error) {
	res := mp.iterAllResults(ctx).CollectAll()

	if err := res.MultiErr(); err != nil {
		return nil, err
	}

	// Сохраняем порядок результатов
	sort.Slice(res.Results, func(i, j int) bool {
		return res.Results[i].Key < res.Results[j].Key
	})

	return res.AvailableResults(), nil
}

// AllResultsOrFirstError дожидается выполнения всех задач или первой возникшей ошибки
// Возвращает или все результаты или первую возникшую ошибку
func (mp MultiPromise[T]) AllResultsOrFirstError(ctx context.Context) ([]T, error) {
	res := mp.iterAllResults(ctx).CollectAllResultsOrFirstError()

	if err := res.MultiErr(); err != nil {
		return nil, err
	}

	// Сохраняем порядок результатов
	sort.Slice(res.Results, func(i, j int) bool {
		return res.Results[i].Key < res.Results[j].Key
	})

	return res.AvailableResults(), nil
}

func (mp MultiPromise[T]) FirstResult(ctx context.Context) (int, T, error) {
	res := mp.iterAllResults(ctx).CollectAllResultsOrFirstError()

	if err := res.MultiErr(); err != nil {
		return 0, std.Zero[T](), err
	}

	if len(res.Results) > 0 {
		return res.Results[0].Key, res.Results[0].Value, nil
	}

	return 0, std.Zero[T](), ErrEmptyMultiPromise
}

func (mp MultiPromise[T]) PartialResult(ctx context.Context) (map[int]T, map[int]error) {
	res := mp.iterResultsUntillCancel(ctx).CollectAll()
	return ResultsWithKeyToMap(res.Results), ErrorsWithKeyToMap(res.Errors)
}
