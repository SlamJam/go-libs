package actors

import (
	"context"
	"sync/atomic"
)

type ActorWithStat[T any] interface {
	Actor

	GetStat() T
}

type actorStat[T any] struct {
	actor

	stat statUpdater[T]
}

type StatReader[T any] interface {
	GetCurrent() T
}

type StatUpdater[T any] interface {
	StatReader[T]

	Update(T)
}

type statUpdater[T any] struct {
	p atomic.Pointer[T]
}

func (st *statUpdater[T]) GetCurrent() T {
	return *st.p.Load()
}

func (st *statUpdater[T]) Update(v T) {
	st.p.Store(&v)
}

func NewActorWithStat[T any](main func(context.Context, StatUpdater[T]) error, opts ...Opt) ActorWithStat[T] {
	a := actorStat[T]{}

	a.actor = *newActor(func(ctx context.Context) error {
		return main(ctx, &a.stat)
	}, opts...)

	var t T
	a.stat.p.Store(&t)

	return &a
}

func (a *actorStat[T]) GetStat() T {
	return a.stat.GetCurrent()
}
