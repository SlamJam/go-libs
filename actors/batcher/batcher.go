package batcher

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/SlamJam/go-libs/actors"
	"github.com/SlamJam/go-libs/options"
	"github.com/SlamJam/go-libs/xchan"
	"github.com/SlamJam/go-libs/xslices"
)

type Batcher[T any] struct {
	actors.Actor

	reuseBuf   bool
	flushCap   int
	flushAfter time.Duration
	outCh      chan []T
	InputCh    chan T
	flushed    atomic.Bool
}

type batcherOptions struct {
	inCap      int
	outCap     int
	reuseBuf   bool
	flushAfter time.Duration
}

type Opt = options.Opt[batcherOptions]

func WithReuseBuffer(reuse bool) Opt {
	return func(o *batcherOptions) {
		o.reuseBuf = reuse
	}
}

func WithInputBuffered(size int) Opt {
	return func(o *batcherOptions) {
		o.inCap = size
	}
}

func WithOutputBuffered(size int) Opt {
	return func(o *batcherOptions) {
		o.outCap = size
	}
}

func WithFlushInterval(interval time.Duration) Opt {
	return func(o *batcherOptions) {
		o.flushAfter = interval
	}
}

func NewBatcher[T any](size int, opts ...Opt) *Batcher[T] {
	var optState batcherOptions

	options.ApplyInto(&optState, opts...)

	b := &Batcher[T]{
		reuseBuf:   optState.reuseBuf,
		flushAfter: optState.flushAfter,
		InputCh:    make(chan T, optState.inCap),
		outCh:      make(chan []T, optState.outCap),
	}

	b.Actor = actors.NewActor(b.do)

	return b
}

func (b *Batcher[T]) C() <-chan []T {
	return b.outCh
}

func (b *Batcher[T]) IsFlushed() bool {
	return b.flushed.Load()
}

func (b *Batcher[T]) do(ctx context.Context) error {
	buf := make([]T, 0, b.flushCap)

	var timer <-chan time.Time
	for {
		var flush bool

		select {
		case item := <-b.InputCh:
			if len(buf) == 0 {
				if b.flushAfter != 0 {
					timer = time.After(b.flushAfter)
				}
				b.flushed.Store(false)
			}

			buf = append(buf, item)
			flush = len(buf) >= b.flushCap
		case <-timer:
			flush = true
		case <-ctx.Done():
			return ctx.Err()
		}

		if flush {
			b.outCh <- buf
			if b.reuseBuf {
				// Dangerous
				buf = buf[:]
			} else {
				buf = xslices.NewWithSameTypeAndCap(buf)
			}

			timer = nil
			b.flushed.Store(true)
		}
	}
}

func Foo() {
	b := NewBatcher[int](1000)

	go func() {
		for batch := range b.C() {
			_ = batch
		}
	}()

	xchan.Put(b.InputCh, 5, 1*time.Second)
}
