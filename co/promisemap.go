package co

import (
	"context"

	std "github.com/SlamJam/go-libs"
	"github.com/SlamJam/go-libs/xmaps"
	"github.com/SlamJam/go-libs/xslices"
	"github.com/pkg/errors"
)

var ErrEmptyPromiseMap = errors.New("empty PromiseMap")

type PromiseMap[K comparable, T any] map[K]Promise[T]

func (pm *PromiseMap[K, T]) Add(k K, f func() (T, error)) {
	pm.Append(k, NewPromise(f))
}

func (pm *PromiseMap[K, T]) Append(k K, p Promise[T]) {
	if *pm == nil {
		*pm = PromiseMap[K, T]{}
	}

	(*pm)[k] = p
}

func (pm PromiseMap[K, T]) AsWaitables() []Awaitable {
	return xslices.Map(xmaps.Values(pm), func(p Promise[T]) Awaitable { return p })
}

func (pm PromiseMap[K, T]) AsPromiseWithKeys() []PromiseWithKey[T, K] {
	result := make([]PromiseWithKey[T, K], 0, len(pm))
	for k, p := range pm {
		result = append(result, PromiseWithKey[T, K]{Key: k, Promise: p})
	}

	return result
}

func (pm PromiseMap[K, T]) iterAllResults(ctx context.Context) Iterator[T, K] {
	return IterAllResults(ctx, pm.AsPromiseWithKeys()...)
}

func (pm PromiseMap[K, T]) iterResultsUntillCancel(ctx context.Context) Iterator[T, K] {
	return IterResultsUntilCancel(ctx, pm.AsPromiseWithKeys()...)
}

func (pm PromiseMap[K, T]) AllResults(ctx context.Context) (map[K]T, error) {
	res := pm.iterAllResults(ctx).CollectAll()

	if err := res.MultiErr(); err != nil {
		return nil, err
	}

	return ResultsWithKeyToMap(res.Results), nil
}

func (pm PromiseMap[K, T]) AllResultsOrFirstError(ctx context.Context) (map[K]T, error) {
	res := pm.iterAllResults(ctx).CollectAllResultsOrFirstError()

	if err := res.MultiErr(); err != nil {
		return nil, err
	}

	return ResultsWithKeyToMap(res.Results), nil
}

func (pm PromiseMap[K, T]) FirstResult(ctx context.Context) (K, T, error) {
	res := pm.iterAllResults(ctx).CollectFirstResult()

	if err := res.MultiErr(); err != nil {
		return std.Zero[K](), std.Zero[T](), err
	}

	if len(res.Results) > 0 {
		return res.Results[0].Key, res.Results[0].Value, nil
	}

	return std.Zero[K](), std.Zero[T](), ErrEmptyMultiPromise
}

func (pm PromiseMap[K, T]) PartialResult(ctx context.Context) (map[K]T, map[K]error) {
	res := pm.iterResultsUntillCancel(ctx).CollectAll()
	return ResultsWithKeyToMap(res.Results), ErrorsWithKeyToMap(res.Errors)
}
