package xmaps

import (
	std "github.com/SlamJam/go-libs"
)

type empty = std.Void

func UniqValues[K comparable, V comparable](m map[K]V) []V {
	result := make([]V, 0, len(m))
	seen := make(map[V]empty, len(result))

	for _, v := range m {
		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = empty{}
		result = append(result, v)
	}

	return result
}

func Keys[K comparable, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	return result
}

func Values[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}

	return result
}

func FromSliceCollect[T any, K comparable, V any](value []T, mapper func(T) (K, V)) map[K][]V {
	result := make(map[K][]V, len(value))
	for _, v := range value {
		k, v := mapper(v)

		result[k] = append(result[k], v)
	}

	return result
}

func FromSlicePrior[T any, K comparable, V any](value []T, mapper func(T) (K, V), less func(i, j int) bool) map[K]V {
	result := make(map[K]V, len(value))
	keyIdx := make(map[K]int, len(value))
	for i, v := range value {
		k, v := mapper(v)

		if j, ok := keyIdx[k]; ok && less(i, j) {
			continue
		}

		result[k] = v
		keyIdx[k] = i
	}

	return result
}

func FromSliceSaveFirst[T any, K comparable, V any](value []T, mapper func(T) (K, V)) map[K]V {
	return FromSlicePrior(value, mapper, func(i, j int) bool { return true })
}

func FromSliceSaveLast[T any, K comparable, V any](value []T, mapper func(T) (K, V)) map[K]V {
	return FromSlicePrior(value, mapper, func(i, j int) bool { return false })
}

func FromSliceFlat[T any, K comparable, V any](value []T, mapper func(T) (K, []V)) map[K][]V {
	result := make(map[K][]V, len(value))
	for _, v := range value {
		k, v := mapper(v)

		result[k] = append(result[k], v...)
	}

	return result
}
