package co

import (
	"context"

	std "github.com/SlamJam/go-libs"
	"github.com/SlamJam/go-libs/xmaps"
)

type PromiseMap[K comparable, T any] map[K]Promise[T]

func (pm *PromiseMap[K, T]) Add(k K, f func() (T, error)) {
	if *pm == nil {
		*pm = PromiseMap[K, T]{}
	}

	(*pm)[k] = NewPromise(f)
}

func (pm PromiseMap[K, T]) multipromise() MultiPromise[T] {
	return xmaps.Values(pm)
}

func (pm PromiseMap[K, T]) results() map[K]T {
	result := make(map[K]T, len(pm))
	for k, p := range pm {
		result[k] = p.Value()
	}

	return result
}

func (pm PromiseMap[K, T]) AllResults(ctx context.Context) (map[K]T, error) {
	_, err := pm.multipromise().AllResults(ctx)
	if err != nil {
		return nil, err
	}

	return pm.results(), nil
}

func (pm PromiseMap[K, T]) AllResultsOrFirstError(ctx context.Context) (map[K]T, error) {
	_, err := pm.multipromise().AllResultsOrFirstError(ctx)
	if err != nil {
		return nil, err
	}

	return pm.results(), nil
}

func (pm PromiseMap[K, T]) FirstResult(ctx context.Context) (K, T, error) {
	_, err := pm.multipromise().FirstResult(ctx)
	if err != nil {
		return std.Zero[K](), std.Zero[T](), err
	}

	// Тут мы точно знаем, что есть хотя бы один не ошибочный готовый результат
	for k, v := range pm {
		if !v.IsCompleted() {
			continue
		}

		if res, err := v.Poll(ctx); err == nil {
			return k, res, nil
		}
	}

	// Такого быть не может
	panic("BUG!!!")
}

type PromiseMapPartialResult[K comparable, T any] struct {
	Results map[K]T
	Errors  map[K]error
}

func (pm PromiseMap[K, T]) PartialResult(ctx context.Context) (result PromiseMapPartialResult[K, T], resultErr error) {
	allResult, err := pm.AllResults(ctx)
	if err == nil {
		result.Results = allResult
		return
	}

	// Пытаемся найти что-то, что успело завершиться
	for k, v := range pm {
		if !v.IsCompleted() {
			continue
		}

		res, err := v.Poll(ctx)

		if err == nil {
			result.Results[k] = res
		} else {
			result.Errors[k] = err
		}
	}

	return
}
