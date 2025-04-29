package actors_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/SlamJam/go-libs/actors"
	"github.com/stretchr/testify/assert"
)

func TestActorInterruption(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	done := make(chan struct{})

	c := actors.NewActor(
		func(ctx context.Context) error {
			<-done
			return nil
		},
	)
	c.Start(ctx)

	err = c.WaitUntilStarted(ctx)
	assert.NoError(t, err)

	fmt.Println(c.IsInterrupted())

	err = c.Interrupt(ctx)
	assert.NoError(t, err)

	assert.Equal(t, true, c.IsInterrupted())

	err = c.Interrupt(ctx)
	if assert.Error(t, err) {
		assert.Equal(t, actors.ErrActorIsNotRunning, err)
	}

	assert.Equal(t, true, c.IsInterrupted())

}

type myActor struct {
	actors.Actor
	Finished bool
}

func NewMyActor() *myActor {
	c := myActor{}
	c.Actor = actors.NewActor(
		func(ctx context.Context) error {
			c.Finished = true
			return nil
		})

	return &c
}

func TestOwnActor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error

	c := NewMyActor()
	c.Start(ctx)
	err = c.WaitUntilHalted(ctx)
	assert.NoError(t, err)

	assert.Equal(t, true, c.Finished)
	assert.Equal(t, true, c.IsHalted())
}
