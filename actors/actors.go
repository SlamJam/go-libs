package actors

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/SlamJam/go-libs/options"
	"github.com/SlamJam/go-libs/xgo"
	"github.com/SlamJam/go-libs/xsync"
	"github.com/pkg/errors"
)

var (
	ErrActorIsNil        = errors.New("actor is nil")
	ErrActorIsNotRunning = errors.New("actor is not running")
)

type actorOptions struct {
	onInterrupt func(context.Context) error
	onHalt      func(context.Context, error)
}

type PanicHandler func(any)

var globalPanicHandler xsync.Value[PanicHandler]

func SetGlobalPanicHandler(h PanicHandler) {
	globalPanicHandler.Store(h)
}

func callGlobalPanicHandler(p any) {
	if h, ok := globalPanicHandler.Load(); ok {
		h(p)
	}
}

// Жизненный цикл:
// Создан -> Запущен -> [Прерван] -> Остановлен/Завершён
type Actor interface {
	Start(ctx context.Context) (result bool)
	Interrupt(ctx context.Context) (err error)

	IsInterrupted() bool
	IsHalted() bool
	IsStarted() bool
	IsRunning() bool

	WaitUntilStarted(ctx context.Context) error
	WaitUntilHalted(ctx context.Context) error

	Error() error
}

type actor struct {
	inited bool

	startOnce  *sync.Once
	cancelOnce *sync.Once
	cancelFunc atomic.Pointer[context.CancelFunc]
	done       chan struct{}
	started    chan struct{}
	haltError  error

	main func(context.Context) error
	opts actorOptions
}

type Opt = options.Opt[actorOptions]

func WithOnInterrupt(f func(context.Context) error) Opt {
	return func(o *actorOptions) {
		o.onInterrupt = f
	}
}

func WithOnHalt(f func(context.Context, error)) Opt {
	return func(o *actorOptions) {
		o.onHalt = f
	}
}

func NewActor(main func(context.Context) error, opts ...Opt) Actor {
	return newActor(main, opts...)
}

func newActor(main func(context.Context) error, opts ...Opt) *actor {
	c := &actor{
		inited: true,

		startOnce:  &sync.Once{},
		cancelOnce: &sync.Once{},
		started:    make(chan struct{}),
		done:       make(chan struct{}),

		main: main,
	}

	options.ApplyInto(&c.opts, opts...)

	return c
}

func (c *actor) mustInitialized() {
	if c == nil {
		panic("actor is nil")
	}

	if !c.inited {
		panic("actor is not initialized. Use .Actor = actors.NewActor(...)")
	}
}

func (c *actor) Start(ctx context.Context) (result bool) {
	c.mustInitialized()

	c.startOnce.Do(func() {
		go func() {
			lctx, cancel := context.WithCancel(ctx)
			c.cancelFunc.Store(&cancel)

			defer func() {
				cancel()
				c.cancelFunc.Store(nil)
				close(c.done)
			}()

			close(c.started)

			c.haltError = xgo.PanicCatcherErr(func() error {
				return errors.WithStack(c.main(lctx))
			})

			if hndl := c.opts.onHalt; hndl != nil {
				hndl(lctx, c.haltError)
			}
		}()

		result = true
	})

	return result
}

func (c *actor) Interrupt(ctx context.Context) (err error) {
	if c == nil {
		return ErrActorIsNil
	}

	c.mustInitialized()

	if cancel := c.cancelFunc.Load(); cancel == nil {
		err = ErrActorIsNotRunning
	} else {
		c.cancelOnce.Do(func() {
			(*cancel)()
			c.cancelFunc.Store(nil)

			if hndl := c.opts.onInterrupt; hndl != nil {
				err = hndl(ctx)
			}
		})
	}

	return err
}

func (c *actor) IsInterrupted() bool {
	return c.cancelFunc.Load() == nil
}

func (c *actor) IsHalted() bool {
	c.mustInitialized()

	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *actor) IsStarted() bool {
	c.mustInitialized()

	select {
	case <-c.started:
		return true
	default:
		return false
	}
}

func (c *actor) IsRunning() bool {
	c.mustInitialized()

	return c.IsStarted() && !c.IsInterrupted()
}

func (c *actor) WaitUntilStarted(ctx context.Context) error {
	c.mustInitialized()

	select {
	case <-c.started:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *actor) WaitUntilHalted(ctx context.Context) error {
	c.mustInitialized()

	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *actor) Error() error {
	return c.haltError
}
