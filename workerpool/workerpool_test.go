package workerpool_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/SlamJam/go-libs/workerpool"
	"github.com/stretchr/testify/assert"
)

func TestWorkerpool(t *testing.T) {
	t.Parallel()

	var total atomic.Uint32

	w := workerpool.New(func(ctx context.Context, i int) {
		total.Add(1)
	}, 8)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Start(ctx)

	err := w.WaitUntilStarted(ctx)
	assert.NoError(t, err)

	for i := range 8 {
		w.Input <- i
	}

	close(w.Input)

	err = w.WaitUntilHalted(ctx)
	assert.NoError(t, err)

	assert.Equal(t, uint32(8), total.Load())
}

func TestWorkerpoolInterrupt(t *testing.T) {
	t.Parallel()

	var total atomic.Uint32

	w := workerpool.New(func(ctx context.Context, i int) {
		<-ctx.Done()
		total.Add(1)
	}, 8)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Start(ctx)

	err := w.WaitUntilStarted(ctx)
	assert.NoError(t, err)

	for i := range 8 {
		w.Input <- i
	}

	err = w.Interrupt(ctx)
	assert.NoError(t, err)

	err = w.WaitUntilHalted(ctx)
	assert.NoError(t, err)

	assert.Equal(t, uint32(8), total.Load())
}
