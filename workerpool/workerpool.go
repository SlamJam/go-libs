package workerpool

import (
	"context"

	"github.com/SlamJam/go-libs/actors"
	"github.com/SlamJam/go-libs/xgo"
	"github.com/SlamJam/go-libs/xsync"
)

type Workerpool interface {
	actors.Actor
}

type Error[T any] struct {
	Item  T
	Error error
}

type Panic[T any] struct {
	Item  T
	Panic any
}

type workerpool[T any] struct {
	actors.Actor

	Input chan T
	// Errors       chan Error[T]
	Panics       chan Panic[T]
	f            func(context.Context, T)
	ctx          context.Context
	maxWorkers   int
	workersCount int
	condr        xsync.RWConditioner
}

func New[T any](f func(context.Context, T), maxworkers int) *workerpool[T] {
	w := &workerpool[T]{
		Input:      make(chan T),
		f:          f,
		maxWorkers: maxworkers,
		condr:      xsync.NewConditionerRW(),
	}

	w.Actor = actors.NewActor(func(ctx context.Context) error {
		w.ctx = ctx
		w.maximiseWorkers()
		w.condr.Wait(func() bool {
			return w.workersCount == 0
		})

		return nil
	})

	return w
}

func (w *workerpool[T]) SetWorkersCount(count int) {
	w.condr.DoAndNotifyAll(func() {
		w.maxWorkers = count
	})

	w.maximiseWorkers()
}

func (w *workerpool[T]) maximiseWorkers() int {
	var result int

	for w.tryStartWorker() {
		result++
	}

	return result
}

func (w *workerpool[T]) isExcess() bool {
	return w.workersCount > w.maxWorkers
}

func (w *workerpool[T]) hasCapacity() bool {
	return w.workersCount < w.maxWorkers
}

func (w *workerpool[T]) tryStartWorker() bool {
	var hasCapacity bool

	w.condr.DoAndNotifyAll(func() {
		hasCapacity = w.hasCapacity()
		if hasCapacity {
			w.workersCount++
		}
	})

	if !hasCapacity {
		return false
	}

	// TODO: GoRecoverForever
	go func() {
		var needEvict bool

		defer func() {
			// Если мы тут, но не по своей воли, вероятно идёт паника.
			// Тогда нужно сделать w.workersCount--
			if !needEvict {
				w.condr.DoAndNotifyAll(func() {
					w.workersCount--
				})
			}
		}()

		for {
			var isExcess bool

			// "дешёвая" проверка
			w.condr.RDo(func() {
				isExcess = w.isExcess()
			})

			if isExcess {
				w.condr.DoAndNotifyAll(func() {
					needEvict = w.isExcess()
					if needEvict {
						w.workersCount--
					}
				})

				if needEvict {
					return
				}
			}

			select {
			case item, ok := <-w.Input:
				if !ok {
					return
				}

				// var err error
				if panicObj := xgo.CatchPanic(func() { w.f(w.ctx, item) }); panicObj != nil {
					w.Panics <- Panic[T]{Item: item, Panic: panicObj}
				}

				// if err != nil {
				// 	w.Errors <- Error[T]{Item: item, Error: err}
				// }
			case <-w.ctx.Done():
				return
			}
		}
	}()

	return true
}
