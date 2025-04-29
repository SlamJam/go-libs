package xchan

import (
	"context"
	"sync"
	"time"

	std "github.com/SlamJam/go-libs"
)

func PutContext[T any](ctx context.Context, ch chan<- T, item T) error {
	select {
	case ch <- item:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func Put[T any](ch chan<- T, item T, d time.Duration) error {
	select {
	case ch <- item:
	case <-time.After(d):
		return std.ErrTimeout
	}

	return nil
}

func IsClosed[T any](ch <-chan T) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func FanIn[T any](chans ...<-chan T) <-chan T {
	result := make(chan T)

	wg := sync.WaitGroup{}
	for _, ch := range chans {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for item := range ch {
				result <- item
			}
		}()
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func Map[IN, OUT any](fn func(IN) OUT) func(<-chan IN) <-chan OUT {
	return func(stream <-chan IN) <-chan OUT {
		result := make(chan OUT)

		go func() {
			defer close(result)

			for item := range stream {
				result <- fn(item)
			}
		}()

		return result
	}
}

func Parallel[IN, OUT any](stream <-chan IN, fn func(IN) OUT, count int) <-chan OUT {
	processStreams := make([]<-chan OUT, count)
	for i := range count {
		processStreams[i] = Map(fn)(stream)
	}

	return FanIn(processStreams...)
}

func Batch[T any](size int) func(stream <-chan T) <-chan []T {
	return func(stream <-chan T) <-chan []T {

		result := make(chan []T)

		go func() {
			defer close(result)

			chunk := make([]T, 0, size)
			closed := false

			for !closed {
				needFlush := false

				select {
				case item, ok := <-stream:
					if !ok {
						closed = true
						needFlush = true
						break
					}

					chunk = append(chunk, item)
					needFlush = len(chunk) >= size
					// case <-timer:
					// needFlush = true
				}

				if needFlush && (len(chunk) > 0) {
					result <- chunk
					chunk = make([]T, 0, size)
				}
			}
		}()

		return result
	}
}

func Drain[IN any](stream <-chan IN) {
	go func() {
		for range stream {
		}
	}()
}

// TrySendNonBlocking пытается отправить значение в канал без блокировки.
// Возвращает true, если значение было отправлено, и false, если канал переполнен.
func TrySendNonBlocking[T any](ch chan T, value T) bool {
	select {
	case ch <- value:
		return true
	default:
		return false
	}
}
