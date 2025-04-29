package xiter

import (
	"iter"
	"slices"

	std "github.com/SlamJam/go-libs"
)

func UniqueWithCap[T any, K comparable](s iter.Seq2[T, K], capacity int) iter.Seq2[T, K] {
	return func(yield func(T, K) bool) {
		seen := make(map[K]std.Void, capacity)
		for i, v := range s {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = std.Void{}
			if !yield(i, v) {
				return
			}
		}
	}
}

func Unique[T any, K comparable](s iter.Seq2[T, K]) iter.Seq2[T, K] {
	return UniqueWithCap(s, 0)
}

func foo() {
	arr := [100_500]int{1, 2, 3}
	s := arr[:]

	for _ = range UniqueWithCap(slices.All(s), 10_000) {

	}
}
