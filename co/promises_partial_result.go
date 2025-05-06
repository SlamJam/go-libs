package co

import "go.uber.org/multierr"

type ResultWithKey[T, K any] struct {
	Key   K
	Value T
}

type ErrorWithKey[K any] struct {
	Key K
	Err error
}

func ResultsWithKeyToMap[K comparable, T any](results []ResultWithKey[T, K]) map[K]T {
	result := make(map[K]T, len(results))
	for _, r := range results {
		result[r.Key] = r.Value
	}
	return result
}

func ErrorsWithKeyToMap[K comparable](errors []ErrorWithKey[K]) map[K]error {
	result := make(map[K]error, len(errors))
	for _, e := range errors {
		result[e.Key] = e.Err
	}
	return result
}

type PartialResult[T, K any] struct {
	Results []ResultWithKey[T, K]
	Errors  []ErrorWithKey[K]
}

func (pr *PartialResult[T, K]) addError(idx K, err error) {
	pr.Errors = append(pr.Errors, ErrorWithKey[K]{Key: idx, Err: err})
}

func (pr *PartialResult[T, K]) addResult(idx K, val T) {
	pr.Results = append(pr.Results, ResultWithKey[T, K]{Key: idx, Value: val})
}

func (pr PartialResult[T, K]) AvailableResults() []T {
	result := make([]T, 0, len(pr.Results))
	for _, r := range pr.Results {
		result = append(result, r.Value)
	}

	return result
}

func (pr PartialResult[T, K]) MultiErr() error {
	var result error
	for _, err := range pr.Errors {
		result = multierr.Append(result, err.Err)
	}

	return result
}
