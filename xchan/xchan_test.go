package xchan_test

import (
	"testing"

	"github.com/SlamJam/go-libs/xchan"
)

func TestExample(t *testing.T) {
	var stream <-chan string

	xchan.Parallel(
		xchan.Batch[string](500)(stream),
		func(batch []string) error {
			return nil
		},
		8,
	)
}
