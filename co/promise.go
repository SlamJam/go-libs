package co

import (
	"context"
	"sync"
	"sync/atomic"

	std "github.com/SlamJam/go-libs"
	"github.com/pkg/errors"
)

// type Deffered[T any] func() (T, error)

type Promise[T any] = *promise[T]

var _ Awaitable = NewPromise(func() (std.Void, error) { return std.Void{}, nil })

type promise[T any] struct {
	f      func() (T, error)
	result T
	err    error

	initialized bool
	launched    atomic.Bool
	comleted    atomic.Bool
	done        chan struct{}

	onceLaunch sync.Once
}

func (p *promise[T]) assertInitialized() {
	if !p.initialized {
		panic("misuse: Promise must be create with NewPromise")
	}
}

func (p *promise[T]) IsLaunched() bool {
	return p.launched.Load()
}

func (p *promise[T]) IsCompleted() bool {
	return p.comleted.Load()
}

func (p *promise[T]) ensureLaunched() (result bool) {
	p.onceLaunch.Do(func() {
		p.launched.Store(true)
		result = true

		go func() {
			defer close(p.done)
			defer p.comleted.Store(true)

			p.result, p.err = p.f()
		}()
	})

	return
}

func (p *promise[T]) Poll(ctx context.Context) (T, error) {
	p.assertInitialized()

	if p.IsCompleted() {
		return p.result, p.err
	}

	p.ensureLaunched()

	select {
	case <-p.done:
		return p.result, p.err

	case <-ctx.Done():
		return std.Zero[T](), ctx.Err()
	}
}

func (p *promise[T]) Await(ctx context.Context) (err error) {
	_, err = p.Poll(ctx)
	return
}

func (p *promise[T]) Value() (result T) {
	if !p.IsCompleted() {
		panic("trying to read a value that is not set")
	}

	result, _ = p.Poll(context.TODO())
	return
}

func newPromise[T any](f func() (T, error), eager bool) Promise[T] {
	p := &promise[T]{
		f: f,

		initialized: true,
		done:        make(chan struct{}),
	}

	if eager {
		p.ensureLaunched()
	}

	return p
}

func NewPromise[T any](f func() (T, error)) Promise[T] {
	return newPromise(f, true)
}

func NewLazyPromise[T any](f func() (T, error)) Promise[T] {
	return newPromise(f, false)
}

func NewResolved[T any](result T) Promise[T] {
	return NewPromise(func() (T, error) {
		return result, nil
	})
}

func NewRejected[T any](err error) Promise[T] {
	return NewPromise(func() (T, error) {
		return std.Zero[T](), err
	})
}

var Resolved = NewResolved(std.Void{})

var ErrRejected = errors.New("rejected")

var Rejected = NewRejected[std.Void](ErrRejected)

var Never = NewPromise(func() (struct{}, error) {
	<-context.Background().Done()
	return struct{}{}, nil
})
