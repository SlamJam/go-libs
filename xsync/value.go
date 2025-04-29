package xsync

import (
	"sync/atomic"
)

type Value[T any] struct {
	v atomic.Value
}

// Load atomically loads and returns the value stored in x.
func (x *Value[T]) Load() (T, bool) {
	v, ok := x.v.Load().(T)
	return v, ok
}

// Store atomically stores val into x.
func (x *Value[T]) Store(val T) {
	x.v.Store(val)
}

// Swap atomically stores new into x and returns the previous value.
func (x *Value[T]) Swap(new T) (old T) {
	return x.v.Swap(new).(T)
}

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return x.v.CompareAndSwap(old, new)
}
