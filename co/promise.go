package co

import (
	"context"
	"sync/atomic"
)

type Deffered[T any] func() (T, error)

type Promise[T any] = *promise[T]

type promise[T any] struct {
	result T
	err    error

	initialized atomic.Bool
	comleted    atomic.Bool
	done        chan struct{}
}

// Сомнения, нежен ли такой публичный метод??
func (p *promise[T]) IsCompleted() bool {
	return p.comleted.Load()
}

func (p *promise[T]) Poll(ctx context.Context) (T, error) {
	if !p.initialized.Load() {
		panic("Promise must be create with AsPromise")
	}

	if p.comleted.Load() {
		return p.result, p.err
	}

	select {
	case <-p.done:
		return p.result, p.err

	case <-ctx.Done():
		var empty T
		return empty, ctx.Err()
	}
}

func (p *promise[T]) Wait(ctx context.Context) (err error) {
	_, err = p.Poll(ctx)
	return
}

func (p *promise[T]) Value() (result T) {
	if !p.comleted.Load() {
		panic("trying to read a value that is not set")
	}

	result, _ = p.Poll(context.TODO())
	return
}

func NewPromise[T any](f func() (T, error)) Promise[T] {
	p := &promise[T]{
		done: make(chan struct{}),
	}

	go func() {
		defer close(p.done)
		defer p.comleted.Store(true)

		p.result, p.err = f()
	}()

	p.initialized.Store(true)

	return p
}

var Resolved = NewPromise(func() (struct{}, error) {
	return struct{}{}, nil
})

var Never = NewPromise(func() (struct{}, error) {
	<-context.Background().Done()
	return struct{}{}, nil
})
