package component_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/SlamJam/go-libs/component"
	"github.com/stretchr/testify/assert"
)

func TestComponentInterruption(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	done := make(chan struct{})

	c := component.NewComponent(
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
		assert.Equal(t, component.ErrComponentIsNotRunning, err)
	}

	assert.Equal(t, true, c.IsInterrupted())

}

type myComponent struct {
	component.Component
	Finished bool
}

func NewMyComponent() *myComponent {
	c := myComponent{}
	c.Component = component.NewComponent(
		func(ctx context.Context) error {
			c.Finished = true
			return nil
		})

	return &c
}

func TestOwnComponent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error

	c := NewMyComponent()
	c.Start(ctx)
	err = c.WaitUntilHalted(ctx)
	assert.NoError(t, err)

	assert.Equal(t, true, c.Finished)
	assert.Equal(t, true, c.IsHalted())
}
