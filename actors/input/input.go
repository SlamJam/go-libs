package input

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrClosed  = errors.New("input closed")
	ErrTimeout = errors.New("timeout reached")
)

type input[T any] struct {
	ch        chan T
	done      chan struct{}
	closeOnce *sync.Once
}

type Pub[T any] interface {
	Ch() chan<- T
	Add(item T) error
	AddWithContext(ctx context.Context, item T) error
	AddWithTimeout(to time.Duration, item T) error
}

type Priv[T any] interface {
	ChIn() <-chan T
	Close()
}

func NewInput[T any]() (Pub[T], Priv[T]) {
	in := input[T]{
		ch:        make(chan T),
		done:      make(chan struct{}),
		closeOnce: &sync.Once{},
	}

	return in, in
}

/* InputPub */

func (in input[T]) Ch() chan<- T {
	return in.ch
}

func (in input[T]) Add(item T) error {
	select {
	case in.ch <- item:
	case <-in.done:
		return ErrClosed
	}

	return nil
}

func (in input[T]) AddWithContext(ctx context.Context, item T) error {
	select {
	case in.ch <- item:
	case <-in.done:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (in input[T]) AddWithTimeout(to time.Duration, item T) error {
	timer := time.NewTimer(to)
	defer timer.Stop()

	select {
	case in.ch <- item:
	case <-in.done:
		return ErrClosed
	case <-timer.C:
		return ErrTimeout
	}

	return nil
}

/* InputPriv */

func (in input[T]) ChIn() <-chan T {
	return in.ch
}

func (in input[T]) Close() {
	in.closeOnce.Do(func() {
		close(in.done)
	})
}
