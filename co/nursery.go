package co

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrCancelled = errors.New("nursery context canceled")
)

type Nursery struct {
	ctx      context.Context
	cancel   context.CancelFunc
	promises Awaitables
	mu       *sync.Mutex
	// isCompleted   *atomic.Bool
	isInitialized bool
}

type NurseryResult struct {
	promises Awaitables
}

func (nr NurseryResult) Await(ctx context.Context) error {
	// n.assertIsInitialized()

	return nr.promises.AwaitAll(ctx)
}

func (n *Nursery) assertIsInitialized() {
	if !n.isInitialized {
		panic("Nursery is not initialized")
	}
}

func (n *Nursery) Ctx() context.Context {
	n.assertIsInitialized()

	return n.ctx
}

// func (n *Nursery) IsCompleted() bool {
// 	n.assertIsInitialized()

// 	return n.isCompleted.Load()
// }

func (n *Nursery) getResult() NurseryResult {
	n.assertIsInitialized()

	return NurseryResult{
		promises: n.promises,
	}
}

func (n *Nursery) onComplete() {
	n.assertIsInitialized()

	n.cancel()
	// n.isCompleted.Store(true)
}

func NewNursery(ctx context.Context) Nursery {
	ctx, cancel := context.WithCancelCause(ctx)

	return Nursery{
		ctx: ctx,
		mu:  &sync.Mutex{},
		// isCompleted:   &atomic.Bool{},
		isInitialized: true,
		cancel:        func() { cancel(ErrCancelled) },
	}
}

func WithContext(ctx context.Context, f func(n Nursery)) NurseryResult {
	n := NewNursery(ctx)

	defer n.onComplete()

	f(n)

	return n.getResult()
}

func WithContextResult[RES any](ctx context.Context, f func(n Nursery) (RES, error)) (NurseryResult, RES, error) {
	n := NewNursery(ctx)

	defer n.onComplete()

	res, err := f(n)
	return n.getResult(), res, err
}

// f - лямбда
func Fork[T any](n Nursery, f func() (T, error)) Promise[T] {
	p := NewPromise(f)

	n.mu.Lock()
	defer n.mu.Unlock()

	n.promises = append(n.promises, p)

	return p
}

func ForkInMultiPromise[T any](n Nursery, mp MultiPromise[T], f func() (T, error)) {
	p := Fork(n, f)
	mp.Append(p)
}

func xxx1() {
	type Foo struct{}
	type Bar struct{}
	type Baz struct{}

	type Result struct {
		Foo Foo
		Baz Baz
	}

	WithContext(context.TODO(), func(n Nursery) {
		var mpFoo MultiPromise[Foo]

		for range 5 {
			ForkInMultiPromise(n, mpFoo, func() (Foo, error) {
				n.Ctx()

				return Foo{}, nil
			})
		}

		pBar := Fork(n, func() (Bar, error) {
			n.Ctx()

			return Bar{}, nil
		})

		pBaz := Fork(n, func() (Baz, error) {
			bar, err := pBar.Poll(n.Ctx())
			if err != nil {
				return Baz{}, err
			}

			_ = bar

			return Baz{}, nil
		})

		foos, err1 := mpFoo.AllResultsOrFirstError(n.Ctx())
		_ = foos
		_ = err1

		baz, err2 := pBaz.Poll(n.Ctx())
		_ = baz
		_ = err2
	})
	// тут n.Ctx() уже будет кенсельнут
}

func xxx2() {
	type Foo struct{}
	type Bar struct{}
	type Baz struct{}

	type Result struct {
		Foo Foo
		Baz Baz
	}

	getFooPromises := func(n Nursery, count int) MultiPromise[Foo] {
		var mpFoo MultiPromise[Foo]

		for range count {
			ForkInMultiPromise(n, mpFoo, func() (Foo, error) {
				n.Ctx()

				return Foo{}, nil
			})
		}

		return mpFoo
	}

	getBarPromise := func(n Nursery) Promise[Bar] {
		return Fork(n, func() (Bar, error) {
			n.Ctx()

			return Bar{}, nil
		})
	}

	getBaz := func(ctx context.Context, bar Bar) (Baz, error) {
		_, _ = ctx, bar

		return Baz{}, nil
	}

	getBazPromise := func(n Nursery, pBar Promise[Bar]) Promise[Baz] {
		return Fork(n, func() (Baz, error) {
			bar, err := pBar.Poll(n.Ctx())
			if err != nil {
				return Baz{}, err
			}

			// получаем Baz, используя Bar
			return getBaz(n.Ctx(), bar)
		})
	}

	WithContext(context.TODO(), func(n Nursery) {
		mpFoo := getFooPromises(n, 5)
		pBar := getBarPromise(n)
		pBaz := getBazPromise(n, pBar)

		foos, err1 := mpFoo.AllResultsOrFirstError(n.Ctx())
		_, _ = foos, err1

		baz, err2 := pBaz.Poll(n.Ctx())
		_, _ = baz, err2
	})
	// тут n.Ctx() уже будет кенсельнут

	_, res, err := WithContextResult(context.TODO(), func(n Nursery) (int, error) {
		mpFoo := getFooPromises(n, 5)
		pBar := getBarPromise(n)
		pBaz := getBazPromise(n, pBar)

		foos, err1 := mpFoo.AllResultsOrFirstError(n.Ctx())
		_, _ = foos, err1

		baz, err2 := pBaz.Poll(n.Ctx())
		_, _ = baz, err2

		return 42, nil
	})

	_, _ = res, err
}

type Response struct {
}

func RequesReplica(context.Context, string) (Response, error) {
	return Response{}, nil
}

func requesShard(n Nursery, addrs []string) Promise[Response] {
	var replicaReqs MultiPromise[Response]

	for _, addr := range addrs {
		ForkInMultiPromise(n, replicaReqs, func() (Response, error) {
			return RequesReplica(n.Ctx(), addr)
		})
	}

	return Fork(n, func() (Response, error) {
		_, resp, err := replicaReqs.FirstResult(n.Ctx())
		return resp, err
	})
}

var ErrReplicaResultTimeout = errors.New("replica time budget exeeded")

func requesShardWithDelay(n Nursery, addrs []string) Promise[Response] {
	return Fork(n, func() (Response, error) {
		var replicaReqs MultiPromise[Response]

		for _, addr := range addrs {
			ForkInMultiPromise(n, replicaReqs, func() (Response, error) {
				return RequesReplica(n.Ctx(), addr)
			})

			waitCtx, cancel := context.WithTimeoutCause(n.Ctx(), 50*time.Millisecond, ErrReplicaResultTimeout)
			defer cancel()

			_, resp, err := replicaReqs.FirstResult(waitCtx)
			if err == nil {
				return resp, nil
			}
		}

		_, resp, err := replicaReqs.FirstResult(n.Ctx())
		return resp, err
	})
}

// P1        | P2        | FirstResult(waitCtx)

// In-Flight | In-Flight | error(ErrReplicaResultTimeout)
// In-Flight | Result    | Result
// In-Flight | Error     | error(ErrReplicaResultTimeout)
// Result    | In-Flight | Result
// Result    | Result    | Result (one of)
// Result    | Error     | Result
// Error     | In-Flight | error(ErrReplicaResultTimeout)
// Error     | Result    | Result
// Error     | Error     | multierr

func xxx3() {
	// shard -> []replica
	cluster := [][]string{
		{"shard1.replica1", "shard1.replica2", "shard1.replica3"},
		{"shard2.replica1", "shard2.replica2", "shard2.replica3"},
		{"shard3.replica1", "shard3.replica2", "shard3.replica3"},
		{"shard4.replica1", "shard4.replica2", "shard4.replica3"},
	}

	nr, resp, err := WithContextResult(context.TODO(), func(n Nursery) ([]Response, error) {
		var shardReqs MultiPromise[Response]
		for _, shard := range cluster {
			p := requesShard(n, shard)
			shardReqs.Append(p)
		}

		// Хотим все результаты
		return shardReqs.AllResults(n.Ctx())

		// Зачем ждать все, если кто-то не ответил?
		// return shardReqs.AllResultsOrFirstError(n.Ctx())

		// Соберём частичный результат
		// partialResult := shardReqs.PartialResult(n.Ctx())
		// if err := partialResult.MultiErr(); err != nil {
		// 	log.Printf("WARN: partial result with errors: %v", err)
		// }

		// return partialResult.AvailableResults(), nil
	})

	_, _ = resp, err

	// Wait until all jobs to be done
	// Wait until shutdown complete
	nr.Await(context.TODO())

	// Focus
	manyNurseryResults := Awaitables{nr, nr, nr}
	_ = manyNurseryResults.AwaitAll(context.TODO())
}
