package co_test

import (
	"context"
	"testing"
	"time"

	"github.com/SlamJam/go-libs/co"
)

func TestPromise(t *testing.T) {
	ctx := context.Background()
	pctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	var mp co.MultiPromise[int]

	mp.Add(func() (int, error) {
		_ = pctx.Done()
		return 10, nil
	})

	mp = append(mp, co.NewPromise(func() (int, error) {
		_ = pctx.Done()
		return 20, nil
	}))

	if res, err := mp.AllResultsOrFirstError(ctx); err != nil {
		t.Log(res)
	}
}
