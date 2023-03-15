package component

import (
	"context"
	"sync"

	"github.com/SlamJam/go-libs/options"
	"github.com/pkg/errors"
)

var (
	ErrComponentIsNil        = errors.New("component is nil")
	ErrComponentIsNotRunning = errors.New("component is not running")
)

type compOptions struct {
	onInterrupt func(context.Context) error
	onHalt      func(context.Context, error)
}

// Жизненный цикл:
// Создан -> Запущен -> [Прерван] -> Остановлен/Завершён
type Component interface {
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

type component struct {
	inited bool

	startOnce  *sync.Once
	cancelOnce *sync.Once
	cancelFunc context.CancelFunc
	done       chan struct{}
	started    chan struct{}
	haltError  error

	main func(context.Context) error
	opts compOptions
}

type Opt = options.Opt[compOptions]

func WithOnInterrupt(f func(context.Context) error) Opt {
	return func(o compOptions) compOptions {
		o.onInterrupt = f
		return o
	}
}

func WithOnHalt(f func(context.Context, error)) Opt {
	return func(o compOptions) compOptions {
		o.onHalt = f
		return o
	}
}

func NewComponent(main func(context.Context) error, opts ...Opt) Component {
	c := &component{
		inited: true,

		startOnce:  &sync.Once{},
		cancelOnce: &sync.Once{},
		started:    make(chan struct{}),
		done:       make(chan struct{}),

		main: main,
	}

	for _, opt := range opts {
		c.opts = opt(c.opts)
	}

	options.ApplyOptsInto(&c.opts, opts...)

	return c
}

func (c *component) mustInitialized() {
	if c == nil {
		panic("component is nil")
	}

	if !c.inited {
		panic("component is not initialized")
	}
}

func (c *component) Start(ctx context.Context) (result bool) {
	c.mustInitialized()

	c.startOnce.Do(func() {
		go func() {
			lctx, cancel := context.WithCancel(ctx)
			c.cancelFunc = cancel

			defer func() {
				cancel()
				c.cancelFunc = nil
				close(c.done)
			}()

			close(c.started)

			c.haltError = errors.WithStack(c.main(lctx))

			if hndl := c.opts.onHalt; hndl != nil {
				hndl(lctx, c.haltError)
			}
		}()

		result = true
	})

	return result
}

func (c *component) Interrupt(ctx context.Context) (err error) {
	if c == nil {
		return ErrComponentIsNil
	}

	c.mustInitialized()

	if cancel := c.cancelFunc; cancel == nil {
		err = ErrComponentIsNotRunning
	} else {
		c.cancelOnce.Do(func() {
			cancel()
			c.cancelFunc = nil

			if hndl := c.opts.onInterrupt; hndl != nil {
				err = hndl(ctx)
			}
		})
	}

	return err
}

func (c *component) IsInterrupted() bool {
	return c.cancelFunc == nil
}

func (c *component) IsHalted() bool {
	c.mustInitialized()

	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *component) IsStarted() bool {
	c.mustInitialized()

	select {
	case <-c.started:
		return true
	default:
		return false
	}
}

func (c *component) IsRunning() bool {
	c.mustInitialized()

	return c.IsStarted() && !c.IsInterrupted()
}

func (c *component) WaitUntilStarted(ctx context.Context) error {
	c.mustInitialized()

	select {
	case <-c.started:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *component) WaitUntilHalted(ctx context.Context) error {
	c.mustInitialized()

	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *component) Error() error {
	return c.haltError
}
