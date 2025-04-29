package xslices

import (
	std "github.com/SlamJam/go-libs"
)

type Mapper[T, K any] func(T) K
type Less func(i, j int) bool

func Identity[T any](item T) T {
	return item
}

func NewWithSameCap[D ~[]T1, S ~[]T2, T1, T2 any](s S) D {
	return make(D, 0, cap(s))
}

func NewWithCapOfLen[D ~[]T1, S ~[]T2, T1, T2 any](s S) D {
	return make(D, 0, len(s))
}

func NewWithSameTypeAndCap[S ~[]T, T any](s S) S {
	return NewWithSameCap[S](s)
}

func NewWithSameTypeAndCapOfLen[S ~[]T, T any](s S) S {
	return NewWithCapOfLen[S](s)
}

// func All[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] {
// 	return func(yield func(int, E) bool) {
// 		for i, v := range s {
// 			if !yield(i, v) {
// 				return
// 			}
// 		}
// 	}
// }

// Unique возвращает слайс уникальных значений.
func Unique[S ~[]T, T comparable](items S) S {
	seen := make(map[T]std.Void, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = std.Void{}
	}

	result := make(S, 0, len(seen))
	for k := range seen {
		result = append(result, k)
	}

	return result
}

// var x = Unique(Map([]int{1, 2, 3}, func(x int) string { return "" }))

// UniqueByKey возвращает массив элементов, уникальных по значению предиката.
// `mapFunc` позволяет получить ключ, по которому определяется уникальность элемента.
func UniqueByKey[S ~[]T, K comparable, T any](s S, mapFunc Mapper[T, K]) S {
	return UniqueMapPrior(
		s, mapFunc, func(i, j int) bool {
			return false
		},
	)
}

// UniqueMapPrior возвращает уникальные значения с наивысшим приоритетом.
// Приоритет задается функцией less, аналогично функции sort.Slice.
// Отобранные значения сохраняют свой порядок в результирующем слайсе.
func UniqueMapPrior[S ~[]T, K comparable, T any](items S, mapper Mapper[T, K], less Less) S {
	idxByKey := make(map[K]int, len(items))
	for i, item := range items {
		key := mapper(item)
		if prevIdx, ok := idxByKey[key]; ok && !less(prevIdx, i) {
			continue
		}
		idxByKey[key] = i
	}

	result := make([]T, 0, len(idxByKey))
	for i, item := range items {
		key := mapper(item)
		if idx, ok := idxByKey[key]; ok && idx == i {
			result = append(result, items[idx])
		}
	}

	return result
}

// Batched разбивает исходный слайс на части размером не более size
func Batched[S ~[]T, T any](items S, size int) []S {
	std.AssertSize(size)

	result := make([]S, 0, (len(items)%size)+1)
	for size < len(items) {
		result = append(result, items[:size:size])
		items = items[size:]
	}

	if len(items) != 0 {
		result = append(result, items[:len(items):len(items)])
	}

	return result
}

func Foo() {
	x := []int{1, 2, 3}
	Batched(x, 2)
}

// Chunked разбивает исходный слайс на count частей примерно равного размера
func Chunked[S ~[]T, T any](items S, count int) []S {
	std.AssertSize(count)

	result := make([]S, 0, count)
	for i := 0; i < count; i++ {
		length := len(items)
		min := i * length / count
		max := ((i + 1) * length) / count

		if min != max {
			result = append(result, items[min:max:max-min])
		}
	}

	return result
}

// UniqueDifference вычитает из first массив second, оставляя только уникальные элементы, используя keyer
// для получения ключа, по которому будет определяться уникальность элементов массива.
func UniqueDifferenceMap[S ~[]T, K comparable, T any](first, second S, keyer Mapper[T, K]) S {
	return UniqueByKey(
		Difference(first, second, keyer),
		keyer,
	)
}

// Difference вычитает из first массив second, используя keyer, как способ определить значение для вычитания.
// `keyer` позволяет получить ключ, по которому будет определяться уникальность элементов массива.
func Difference[S ~[]T, K comparable, T any](first, second S, keyer Mapper[T, K]) S {
	smaller, other := first, second

	if len(first) > len(second) {
		smaller, other = second, first
	}

	result := make([]T, 0, len(smaller))
	seen := make(map[K]std.Void, len(smaller))

	for _, item := range smaller {
		seen[keyer(item)] = std.Void{}
	}

	for _, item := range other {
		if _, found := seen[keyer(item)]; found {
			continue
		}
		result = append(result, item)
	}

	return result
}

// Map преобразует коллекцию items из []T в []V с помощью `mapper(T) V`
func Map[T, V any](items []T, mapper Mapper[T, V]) []V {
	result := make([]V, 0, len(items))
	for _, item := range items {
		result = append(result, mapper(item))
	}

	return result
}

// func All[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] {
// 	return func(yield func(int, E) bool) {
// 		for i, v := range s {
// 			if !yield(i, v) {
// 				return
// 			}
// 		}
// 	}
// }

// MapE преобразует коллекцию items из T в V с помощью `mapper` с возвратом ошибки в случае неудачи
func MapE[T, V any](items []T, mapper func(T) (V, error)) ([]V, error) {
	result := make([]V, 0, len(items))
	for _, item := range items {
		mapped, err := mapper(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}

// Each выполняет функцию `fn` для каждого элемента
func Each[T any](items []T, fn func(T, int)) {
	for i, v := range items {
		fn(v, i)
	}
}

// EachE выполняет функцию `fn` для каждого элемента. В случае ошибки возвращает error и индекс элемента
func EachE[T any](items []T, fn func(T, int) error) (error, int) {
	for i, v := range items {
		if err := fn(v, i); err != nil {
			return err, i
		}
	}

	return nil, 0
}

func SplitAt[S ~[]T, T any](items S, size int) (S, S) {
	std.AssertSize(size)

	end := size
	if end > len(items) {
		end = len(items)
	}

	head, tail := items[:end:end], items[end:len(items):len(items)-end]
	return head, tail
}

func TailOf[S ~[]T, T any](items S, size int) S {
	std.AssertSize(size)

	start := len(items) - size

	if start < 0 {
		start = 0
	}

	tail := items[start : len(items) : len(items)-start]
	return tail
}
